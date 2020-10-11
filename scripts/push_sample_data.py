#!/usr/bin/python3
import argparse
import base64
import random
import string
import sys
from datetime import datetime, timedelta
from typing import List, Union

import requests
from tqdm import tqdm

UA = 'wakatime/13.0.7 (Linux-4.15.0-91-generic-x86_64-with-glibc2.4) Python3.8.0.final.0 vscode/1.42.1 vscode-wakatime/4.0.0'
LANGUAGES = {
    'Go': 'go',
    'Java': 'java',
    'JavaScript': 'js',
    'Python': 'py'
}


class Heartbeat:
    def __init__(
            self,
            entity: str,
            project: str,
            language: str,
            time: float,
            is_write: bool = True,
            branch: str = 'master',
            type: str = 'file'
    ):
        self.entity: str = entity
        self.project: str = project
        self.language: str = language
        self.time: float = time
        self.is_write: bool = is_write
        self.branch: str = branch
        self.type: str = type
        self.category: Union[str, None] = None


def generate_data(n: int, n_projects: int = 5, n_past_hours: int = 24) -> List[Heartbeat]:
    data: List[Heartbeat] = []
    now: datetime = datetime.today()
    projects: List[str] = [randomword(random.randint(5, 10)) for _ in range(n_projects)]
    languages: List[str] = list(LANGUAGES.keys())

    for _ in range(n):
        p: str = random.choice(projects)
        l: str = random.choice(languages)
        f: str = randomword(random.randint(2, 8))
        delta: timedelta = timedelta(
            hours=random.randint(0, n_past_hours - 1),
            minutes=random.randint(0, 59),
            seconds=random.randint(0, 59)
        )

        data.append(Heartbeat(
            entity=f'/home/me/dev/{p}/{f}.{LANGUAGES[l]}',
            project=p,
            language=l,
            time=(now - delta).timestamp()
        ))

    return data


def post_data_sync(data: List[Heartbeat], url: str, api_key: str):
    encoded_key: str = str(base64.b64encode(api_key.encode('utf-8')), 'utf-8')

    for h in tqdm(data):
        r = requests.post(url, json=[h.__dict__], headers={
            'User-Agent': UA,
            'Authorization': f'Basic {encoded_key}'
        })
        if r.status_code != 201:
            print(r.text)
            sys.exit(1)


def randomword(length: int) -> str:
    letters = string.ascii_lowercase
    return ''.join(random.choice(letters) for _ in range(length))


def parse_arguments():
    parser = argparse.ArgumentParser(description='Wakapi test data insertion script.')
    parser.add_argument('-n', type=int, default=20, help='total number of random heartbeats to generate and insert')
    parser.add_argument('-u', '--url', type=str, default='http://localhost:3000/api/heartbeat',
                        help='url of your api\'s heartbeats endpoint')
    parser.add_argument('-k', '--apikey', type=str, required=True,
                        help='your api key (to get one, go to the web interface, create a new user, log in and copy the key)')
    parser.add_argument('-p', '--projects', type=int, default=5, help='number of different fake projects to generate')
    parser.add_argument('-o', '--offset', type=int, default=24,
                        help='negative time offset in hours from now for to be used as an interval within which to generate heartbeats for')
    return parser.parse_args()


if __name__ == '__main__':
    args = parse_arguments()

    data: List[Heartbeat] = generate_data(args.n, args.projects, args.offset)
    post_data_sync(data, args.url, args.apikey)
