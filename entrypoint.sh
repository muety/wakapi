#!/bin/bash

if [ "$WAKAPI_DB_TYPE" == "sqlite3" ] || [ "$WAKAPI_DB_TYPE" == "" ]; then
  ./wakapi
else
  echo "Waiting for database to come up"
  ./wait-for-it.sh   "$WAKAPI_DB_HOST:$WAKAPI_DB_PORT" -s -t 60 -- ./wakapi
fi