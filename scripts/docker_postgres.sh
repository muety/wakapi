#!/bin/bash

docker volume create wakapi-postgres-data
docker run -d -p 5432:5432 -e POSTGRES_DB=wakapi_local -e POSTGRES_USER=wakapi_user -e POSTGRES_PASSWORD=wakapi -v wakapi-postgres-data:/var/lib/postgresql --name wakapi-postgres postgres:18-alpine