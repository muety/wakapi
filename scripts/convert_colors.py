#!/usr/bin/env python3

"""
usage: convert_colors.py [-h] [--languages | --no-languages] [--editors | --no-editors]
                         [--operating_systems | --no-operating_systems]

options:
  -h, --help            show this help message and exit
  --languages, --no-languages
  --editors, --no-editors
  --operating_systems, --no-operating_systems
"""

import argparse
import json
import urllib.request

from lxml import html as lxml_html  # pip install lxml


def fetch(url):
    with urllib.request.urlopen(url) as f:
        return f.read().decode('utf-8')


def fetch_wakatime_website(section):
    tree = lxml_html.fromstring(fetch(f'https://wakatime.com/colors/{section}'))
    return {el.attrib['title']: el.xpath('./div[1]/text()')[0].strip()
            for el in tree.xpath('//span[@class="editor-icon tip"]')}


def main():
    parser = argparse.ArgumentParser()
    for s in ('languages', 'editors', 'operating_systems'):
        parser.add_argument(f'--{s}', action=argparse.BooleanOptionalAction, default=True)
    args = parser.parse_args()

    result = {}
    if args.languages:
        result['languages'] = {k: v['color'] for k, v in json.loads(fetch('https://raw.githubusercontent.com/ozh/github-colors/master/colors.json')).items()}
    if args.editors:
        result['editors'] = fetch_wakatime_website('editors')
    if args.operating_systems:
        result['operating_systems'] = fetch_wakatime_website('operating_systems')

    print(json.dumps(result, indent=4))


if __name__ == '__main__':
    main()
