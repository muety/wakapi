#!/bin/bash

# Script to create a new user with Wakapi and return its API key
#
# Example usage:
# ./create_user.sh <USERNAME> <PASSWORD> [<EMAIL>] > my_api_key.txt
#
# By default, this script will run against http://localhost:3000,
# however, you can change the API URL by setting WAKAPI_API_URL environment variable.

WAKAPI_API_URL=${WAKAPI_API_URL:-http://localhost:3000}

echo "⏳ Creating new user ..." >&2
curl -s --request POST \
  --url $WAKAPI_API_URL/signup \
  --fail \
  --data username=$1 \
  --data password=$2 \
  --data password_repeat=$2 \
  --data email=$3 > /dev/null

if [ $? -eq 0 ]; then
  echo "✅ Successfully created user" >&2
else
  echo "❌ Failed to create user" >&2
  exit 1
fi

echo "⏳ Logging in ..." >&2

curl -s \
  --request POST \
  --url $WAKAPI_API_URL/login \
  --fail \
  --cookie-jar /tmp/.wakapi_cookies.txt \
  --data username=$1 \
  --data password=$2

if [ $? -eq 0 ]; then
  echo "✅ Successfully logged in" >&2
else
  echo "❌ Failed to log in" >&2
  exit 1
fi

echo "⏳ Fetching dashboard page and extracting key..." >&2

curl -s \
  --url "$WAKAPI_API_URL/summary?interval=today" \
  --fail \
  --cookie /tmp/.wakapi_cookies.txt | grep -oP --max-count 1 '(.{8}-.{4}-.{4}-.{4}-.{12})'

if [ $? -gt 0 ]; then
  echo "❌ Failed to fetch dashboard" >&2
  exit 1
fi

rm /tmp/.wakapi_cookies.txt