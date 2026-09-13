# convert colors from the format provided by https://raw.githubusercontent.com/ozh/github-colors/master/colors.json to what wakapi wants

import json
import urllib.request

if __name__ == '__main__':
    with urllib.request.urlopen('https://raw.githubusercontent.com/ozh/github-colors/master/colors.json') as f:
        colors = json.load(f)

    result = {k: v['color'] for k, v in colors.items()}
    print(json.dumps(result, indent=4))
