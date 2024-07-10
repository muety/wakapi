#!/bin/bash

echo "Starting smtp4dev in Docker ..."
docker run -d --rm -p 2525:25 -p 8080:80 --name smtp4dev_wakapi rnwood/smtp4dev

echo "Running tests ..."
go test -run TestSmtpTestSuite ../services/mail

echo "Killing container again ..."
docker stop smtp4dev_wakapi