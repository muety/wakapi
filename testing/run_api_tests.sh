#!/bin/bash
set -e

if ! command -v newman &> /dev/null
then
    echo "Newman could not be found. Run 'npm install -g newman' first."
    exit 1
fi

script_path=$(realpath "${BASH_SOURCE[0]}")
script_dir=$(dirname "$script_path")

echo "Compiling."
(cd "$script_dir/.." || exit 1; CGO_ENABLED=0 go build)

cd "$script_dir" || exit 1

# Download previous release (when upgrade testing)
INITIAL_RUN_EXE="../wakapi"
if [[ "$2" == "y" ]]; then
    if [ ! -f wakapi_linux_amd64.zip ]; then
        curl https://github.com/muety/wakapi/releases/latest/download/wakapi_linux_amd64.zip -O -L
    fi
    unzip -o wakapi_linux_amd64.zip
    INITIAL_RUN_EXE="./wakapi"
    echo "Running tests with release version"
fi

# Initialise test data
case $1 in
    sqlite|*)
    rm -f wakapi_testing.db

    echo "Creating database and schema ..."
    sqlite3 wakapi_testing.db < schema.sql

    echo "Importing seed data ..."
    sqlite3 wakapi_testing.db < data.sql

    CONFIG="config.testing.yml"
    ;;
esac

wait_for_wakapi () {
    counter=0
    echo "Waiting for Wakapi to come up ..."
    until curl --output /dev/null --silent --get --fail http://localhost:3000/api/health; do
        if [ "$counter" -ge 5 ]; then
            echo "Waited for 5s, but Wakapi failed to come up ..."
            exit 1
        fi

        printf '.'
        sleep 1
        counter=$((counter+1))
    done
    printf "\n"
}

# Run tests
echo "Running Wakapi testing instance in background ..."
"$INITIAL_RUN_EXE" -config "$CONFIG" &
pid=$!

echo "Running test collection ..."
newman run "wakapi_api_tests.postman_collection.json"
exit_code=$?

echo "Shutting down Wakapi ..."
kill -TERM $pid

# Run upgrade tests
if [[ "$2" == "y" ]]; then
    echo "Now running migrations with build"
    ../wakapi -config "$CONFIG" &
    pid=$!

    wait_for_wakapi
    echo "Shutting down Wakapi ..."
    kill -TERM $pid
fi

echo "Exiting with status $exit_code"
exit $exit_code
