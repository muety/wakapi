import argparse
import time

import requests
from tqdm import tqdm
from typing import List, Tuple, Any, Dict

http: requests.Session = requests.Session()


def read_recipients(file_path: str) -> List[str]:
    with open(file_path, 'r') as f:
        return [r.strip() for r in f.readlines()]


def read_mail_content(file_path: str) -> Tuple[str, bool]:
    with open(file_path, 'r') as f:
        content = f.read()
    return content, file_path.lower().endswith('.html')


def send(recipient: str, subject: str, content: str, is_html: bool = False, base_url: str = 'https://mailwhale.dev'):
    payload: Dict[str, Any] = {
        'to': [recipient],
        'subject': subject,
    }
    if is_html:
        payload['html'] = content
    else:
        payload['text'] = content

    r = http.post(f'{base_url}/api/mail', json=payload)
    r.raise_for_status()


def run(args):
    content, is_html = read_mail_content(args.content)
    for recipient in tqdm(read_recipients(args.recipients)):
        send(recipient, args.subject, content, is_html, args.mw_url)
        time.sleep(args.throttle)


def parse_arguments():
    parser = argparse.ArgumentParser(description='Script to send mass mail to a list of recipients through MailWhale')
    parser.add_argument('-r', '--recipients', required=True, type=str, help='path to line-separated file containing list of recipient addresses')
    parser.add_argument('-c', '--content', required=True, type=str, help='path to text- or html file containing the mail content')
    parser.add_argument('-s', '--subject', required=True, type=str, help='mail subject')
    parser.add_argument('--mw_client_id', required=True, type=str, help='mailwhale client id')
    parser.add_argument('--mw_client_secret', required=True, type=str, help='mailwhale client secret')
    parser.add_argument('--mw_url', default='https://mailwhale.dev', type=str, help='mailwhale base url')
    parser.add_argument('-t', '--throttle', default=5, type=int, help='seconds to wait between every mail')
    return parser.parse_args()


if __name__ == '__main__':
    args = parse_arguments()
    http.auth = (args.mw_client_id, args.mw_client_secret)
    run(args)
