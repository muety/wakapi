#!/bin/bash

if [ "$WAKAPI_DB_TYPE" != "sqlite3" ]; then
  echo "Waiting 10 Seconds for DB to start"
  sleep 10;
fi

echo "Starting Application"
./wakapi