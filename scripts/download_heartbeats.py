#!/bin/python

# Setup:
# pip install requests tqdm

import argparse
import base64
import csv
import datetime
import logging
import os.path
import sys
from typing import Tuple, Any, Dict, List, Generator
from urllib.parse import urlparse

from tqdm import tqdm

import requests

# globals
http: requests.Session = requests.Session()
api_url: str = 'https://wakapi.dev/api'

Heartbeats = List[Dict[str, Any]]

model_keys: List[str] = [
    'branch', 'category', 'entity', 'is_write', 'language', 'project', 'time', 'type', 'user_id', 'machine_name_id', 'user_agent_id', 'created_at', 'lines', 'lineno', 'cursorpos', 'line_deletions', 'line_additions'
]


def fetch_total_range(min_date: datetime.date, max_date: datetime.date) -> Tuple[datetime.date, datetime.date]:
    r = http.get(f'{api_url}/compat/wakatime/v1/users/current/all_time_since_today')
    r.raise_for_status()
    data = r.json()
    start_date: datetime.date = datetime.date.fromisoformat(data['data']['range']['start_date'])
    end_date: datetime.date = datetime.date.fromisoformat(data['data']['range']['end_date'])
    return max([start_date, min_date]), min([end_date, max_date])


def fetch_all_heartbeats(start: datetime.date, end: datetime.date) -> Generator[Heartbeats, None, None]:
    date_range: List[datetime.date] = [start + datetime.timedelta(days=x) for x in range(0, (end - start).days + 2)]
    for date in tqdm(date_range):
        yield fetch_heartbeats(date)


def fetch_heartbeats(date: datetime.date) -> Heartbeats:
    r = http.get(f'{api_url}/compat/wakatime/v1/users/current/heartbeats', params=dict(date=date.isoformat()))
    r.raise_for_status()
    return r.json()['data']


def run(from_date: datetime.date, to_date: datetime.date, out_file: str):
    total_range: Tuple[datetime.date, datetime.date] = fetch_total_range(from_date, to_date)

    with open(out_file, 'w') as f:
        w: csv.DictWriter = csv.DictWriter(f, model_keys, extrasaction='ignore')
        w.writeheader()

        for heartbeats in fetch_all_heartbeats(*total_range):
            for hb in heartbeats:
                w.writerow(hb)


def parse_args():
    parser = argparse.ArgumentParser(description='Script to download raw heartbeats from Wakapi API')
    parser.add_argument('--api_key', required=True, help='Wakapi API key')
    parser.add_argument('--url', required=False, default=api_url, help='Wakapi instance API URL (without trailing slash)')
    parser.add_argument('--from', default='1970-01-01', type=datetime.date.fromisoformat, help='Date range start')
    parser.add_argument('--to', default=datetime.date.today().isoformat(), type=datetime.date.fromisoformat, help='Date range end')
    parser.add_argument('--output', '-o', default='wakapi_heartbeats.csv', help='CSV file to save heartbeats to')
    return parser.parse_args()


def init_http(api_key: str):
    global api_url
    encoded_key: str = str(base64.b64encode(api_key.encode('utf-8')), 'utf-8')
    http.headers.update({'Authorization': f'Basic {encoded_key}'})
    api_url = args.url


if __name__ == '__main__':
    args = parse_args()

    if os.path.exists(args.output):
        logging.error('output file already existing, please delete or choose a different path')
        sys.exit(1)

    if 'wakatime.com' in urlparse(args.url):
        logging.warning('warning: this script is not perfectly compatible with wakatime')

    init_http(args.api_key)
    run(getattr(args, 'from'), getattr(args, 'to'), args.output)
