#!/bin/bash

# Requires Go and Docker to be installed
# Run once initially: go get github.com/mattn/go-sqlite3

VERSION=$(cat version.txt)

xgo -targets linux/amd64,darwin/amd64,windows/amd64 -dest build -out "wakapi_$VERSION" .