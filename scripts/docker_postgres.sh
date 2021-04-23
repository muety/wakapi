#!/bin/bash

docker run -d -p 5432:5432 -e POSTGRES_DB=wakapi_local -e POSTGRES_USER=wakapi_user -e POSTGRES_PASSWORD=wakapi --name wakapi-postgres postgres