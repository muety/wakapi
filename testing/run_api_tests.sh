#!/bin/bash

if [ ! -f "wakapi" ]; then
    echo "Wakapi executable not found. Run 'go build' first."
    exit 1
fi

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
screen -S wakapi_testing -dm bash -c "../wakapi -config config.testing.yml"

echo "Waiting for Wakapi to come up ..."
until $(curl --output /dev/null --silent --get --fail http://localhost:3000/api/health); do
    printf '.'
    sleep 1
done

echo ""

echo "Running test collection ..."
newman run "Wakapi API Tests.postman_collection.json"

echo "Shutting down Wakapi ..."
screen -S wakapi_testing -X quit

echo "Deleting database ..."
rm wakapi_testing.db