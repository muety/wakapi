#!/bin/bash

if [ "$WAKAPI_DB_TYPE" != "sqlite3" ]; then
  echo "Waiting 3 Seconds for DB to start"
  sleep 3;
fi

echo "Starting Application"
./wakapi
