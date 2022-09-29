#!/bin/bash

echo "Compiling."
CGO_ENABLED=0 go build

if ! command -v newman &> /dev/null
then
    echo "Newman could not be found. Run 'npm install -g newman' first."
    exit 1
fi

cd "$(dirname "$0")"

echo "Creating database and schema ..."
sqlite3 wakapi_testing.db < schema.sql

echo "Importing seed data ..."
sqlite3 wakapi_testing.db < data.sql

echo "Running Wakapi testing instance in background ..."
../wakapi -config config.testing.yml &
pid=$!

echo "Waiting for Wakapi to come up ..."
until $(curl --output /dev/null --silent --get --fail http://localhost:3000/api/health); do
    printf '.'
    sleep 1
done

echo ""

echo "Running test collection ..."
newman run "wakapi_api_tests.postman_collection.json"
exit_code=$?

echo "Shutting down Wakapi ..."
kill -TERM $pid

echo "Deleting database ..."
rm wakapi_testing.db

exit $exit_code
