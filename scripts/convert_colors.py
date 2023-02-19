# convert colors from the format provided by https://raw.githubusercontent.com/ozh/github-colors/master/colors.json to what wakapi wants

import sys
import json

with open(sys.argv[1], 'r') as f:
    colors = json.load(f)

result = {}
for k, v in colors.items():
    result[k] = v['color']

print(json.dumps(result, indent=4))
