#!/bin/bash

docker run -d -p 3306:3306 -e MARIADB_ROOT_PASSWORD=secretpassword -e MARIADB_DATABASE=wakapi_local -e MARIADB_USER=wakapi_user -e MARIADB_PASSWORD=wakapi --name wakapi-mariadb mariadb:latest