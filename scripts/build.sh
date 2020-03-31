#!/bin/bash

OSLIST=( darwin linux windows )
ARCHLIST=( amd64 )
VERSION=$(cat version.txt)

for os in ${OSLIST[*]}
  do
    for arch in ${ARCHLIST[*]}
      do
        echo "Building $os / $arch"
        GOOS=$os GOARCH=$arch go build -o "build/wakapi_${VERSION}_${os}_${arch}" "github.com/muety/wakapi"
      done
  done