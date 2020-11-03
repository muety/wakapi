#!/bin/bash

docker run -d -p 3306:3306 -e MYSQL_ROOT_PASSWORD=secretpassword -e MYSQL_DATABASE=wakapi_local -e MYSQL_USER=wakapi_user -e MYSQL_PASSWORD=wakapi --name wakapi-mysql mysql:5