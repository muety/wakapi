#!/bin/bash

if [ "$WAKAPI_DB_TYPE" == "sqlite3" ] || [ "$WAKAPI_DB_TYPE" == "" ]; then
  exec ./wakapi
else
  echo "Waiting for database to come up"
  exec ./wait-for-it.sh   "$WAKAPI_DB_HOST:$WAKAPI_DB_PORT" -s -t 60 -- ./wakapi
fi
