#!/bin/bash

# Load testing tool for Wakapi, requesting summaries at random intervals within the past 6 months.
# Caution: you must not run this against wakapi.dev production instance!
# Usage: ./loaddtest.sh --api-url <url> --api-key <key> [-n <num_requests>] [-c <num_clients>] [--recompute]

# Requirements:
# - https://github.com/hatoo/oha

# Defaults
API_KEY=""
API_URL=""
NUM_REQUESTS=1000
NUM_CLIENTS=10
RECOMPUTE=0

# Parse parameters
OPTIONS=$(getopt -o n:c: --long api-key:,api-url:,recompute -n "$0" -- "$@")
eval set -- "$OPTIONS"

while true; do
  case $1 in
    --api-key) API_KEY="$2"; shift 2;;
    --api-url) API_URL="$2"; shift 2;;
    -n) NUM_REQUESTS="$2"; shift 2;;
    -c) NUM_CLIENTS="$2"; shift 2;;
    --recompute) RECOMPUTE=1; shift;;
    --) shift; break;;
    *) echo "Invalid option: $1"; exit 1;;
  esac
done

# Validate parameters
if [ -z "$API_KEY" ]; then
  echo "API key is required"
  exit 1
fi

if [ -z "$API_URL" ]; then
  echo "API URL is required"
  exit 1
fi

# Runtime variables
TOKEN="$(echo $API_KEY | base64)"
TO_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
FROM_DATE_PREFIX=$(date -u -d "-6 months" +"%Y-%m")
RECOMPUTE_PARAM=$(if [ $RECOMPUTE -eq 1 ]; then echo "true"; else echo "false"; fi)
URL_PATTERN="$API_URL/summary\?from=$FROM_DATE_PREFIX-(?:0[1-9]|1[0-9]|2[0-8])T(?:[01][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9]Z&to=$TO_DATE&recompute=$RECOMPUTE_PARAM"

# Run the test!
oha -n $NUM_REQUESTS -c $NUM_CLIENTS \
  -H "Authorization: Bearer $TOKEN" \
  --disable-keepalive \
  --latency-correction \
  --rand-regex-url $URL_PATTERN