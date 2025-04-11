#!/bin/bash
set -o nounset -o errexit -o pipefail

cleanup() {
    echo "Stopping and removing existing smtp4dev instances ..."
    docker stop smtp4dev_wakapi &> /dev/null || true
    docker rm -f smtp4dev_wakapi &> /dev/null
}
trap cleanup EXIT

cleanup

echo "Starting smtp4dev in Docker ..."
docker run -d --rm -p 2525:25 -p 8080:80 --name smtp4dev_wakapi rnwood/smtp4dev

echo "Running tests ..."
script_dir=$(dirname "${BASH_SOURCE[0]}")
go test -count=1 -run TestSmtpTestSuite "$script_dir/../services/mail"
