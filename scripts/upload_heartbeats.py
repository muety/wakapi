#!/bin/python

# Script to insert Wakapi heartbeats from a CSV file via API.
# Please note, though:
# Recommended way for migrating data between remote and local (or second remote) instance is to use the "WakaTime Import" feature instead (see Settings -> Integrations).
# Recommended way for migrating data between two local / self-hosted instances is to simply copy the "heartbeats" database table.

# Setup:
# pip install requests tqdm pandas

import argparse
import base64
import json
import logging
import time
from typing import List, Dict, Any, Iterator
from urllib.parse import urlparse

import numpy as np
import pandas as pd
import requests
from tqdm import tqdm

# globals
http: requests.Session = requests.Session()
api_url: str = 'https://wakapi.dev/api'


def send_batch(data: List[Dict[str, Any]]):
    user_agent = data[0]['user_agent']  # all user agents in this batch must be identical
    machine = data[0]['machine']  # all machine names in this match must be identical

    r = http.post(f'{api_url}/heartbeats', json=data, headers={
        'User-Agent': user_agent,
        'X-Machine-Name': machine,
    })
    r.raise_for_status()


def split_chunks(df: pd.DataFrame, max_n: int) -> List[pd.DataFrame]:
    n_chunks = (len(df) // max_n) + 1
    return np.array_split(df, n_chunks)


def split_batches(df: pd.DataFrame, max_n: int = 250) -> Iterator[pd.DataFrame]:
    df_grouped = df.groupby(['user_agent', 'machine'])
    for k in df_grouped.groups.keys():
        df = df_grouped.get_group(k)
        for df in split_chunks(df, max_n):
            yield df


def run(file: str, batch_size: int):
    df = pd.read_csv(file)
    df = df.rename(columns={'machine_name_id': 'machine', 'user_agent_id': 'user_agent'})
    df = df.drop(columns=['created_at'])

    for df in tqdm(list(split_batches(df, batch_size))):  # must collect iterator back into list so we can do tqdm :-/
        df_json = json.loads(df.to_json(orient='records'))
        send_batch(df_json)


def parse_args():
    parser = argparse.ArgumentParser(description='Script to upload raw heartbeats from CSV (not recommended)')
    parser.add_argument('--file', '-f', required=True, help='CSV file to read heartbeats from')
    parser.add_argument('--api_key', required=True, help='Wakapi API key')
    parser.add_argument('--url', required=False, default=api_url, help='Wakapi instance API URL (without trailing slash)')
    parser.add_argument('--batch', default=250, help='Maximum size for batched inserts')
    return parser.parse_args()


def init_http(args):
    global api_url
    encoded_key: str = str(base64.b64encode(args.api_key.encode('utf-8')), 'utf-8')
    http.headers.update({'Authorization': f'Basic {encoded_key}'})
    api_url = args.url


if __name__ == '__main__':
    args = parse_args()

    if 'wakatime.com' in urlparse(args.url):
        logging.warning('warning: this script is not perfectly compatible with wakatime')

    logging.warning('Please note: Wakapi will only permit heartbeats up to a certain age (1 week by default). If data from your CSV is older, you will have to adapt "heartbeat_max_age" on the destination server accordingly.')
    time.sleep(3)

    init_http(args)
    run(args.file, args.batch)
