#!/bin/bash

OSLIST=( darwin linux windows )
ARCHLIST=( amd64 )
VERSION=$(cat version.txt)

for os in ${OSLIST[*]}
  do
    for arch in ${ARCHLIST[*]}
      do
        GOOS=$os
        GOARCH=$arch
        echo "Building $GOOS / $GOARCH"
        GOOS=$GOOS GOARCH=$GOARCH go build -o "build/wakapi_${VERSION}_${GOOS}_${GOARCH}" "github.com/muety/wakapi"
      done
  done