#!/bin/bash

if ! command -v newman &> /dev/null
then
    echo "Newman could not be found. Run 'npm install -g newman' first."
    exit 1
fi

for i in "$@"; do
    case $i in
        --migration)
            MIGRATION=1
            ;;
    esac
done

script_path=$(realpath "${BASH_SOURCE[0]}")
script_dir=$(dirname "$script_path")

echo "Compiling."
(cd "$script_dir/.." || exit 1; CGO_ENABLED=0 go build)

cd "$script_dir" || exit 1

# Download previous release (when upgrade testing)
initial_run_exe="../wakapi"
if [[ $MIGRATION -eq 1 ]]; then
    if [ ! -f wakapi_linux_amd64.zip ]; then
        echo "Downloading latest release"
        curl https://github.com/muety/wakapi/releases/latest/download/wakapi_linux_amd64.zip -O -L
    fi
    unzip -o wakapi_linux_amd64.zip
    initial_run_exe="./wakapi"
    echo "Running tests with release version"
fi

cleanup() {
    if [ -n "$pid" ] && ps -p "$pid" > /dev/null; then
        kill -TERM "$pid"
    fi
    if [ "${docker_down-0}" -eq 1 ]; then
        docker compose -f "$script_dir/docker-compose.yml" down
    fi
}
trap cleanup EXIT

# Initialise test data
case $1 in
    postgres|mysql|mariadb)
    docker compose -f "$script_dir/docker-compose.yml" down
    docker volume rm "testing_wakapi-$1"

    docker_down=1
    docker compose -f "$script_dir/docker-compose.yml" up --wait -d "$1"

    config="config.$1.yml"
    if [ "$1" == "mariadb" ]; then
        config="config.mysql.yml"
    fi

    db_port=0
    if [ "$1" == "postgres" ]; then
        db_port=55432
    else
        db_port=53306
    fi

    for _ in $(seq 0 30); do
        if netstat -tulpn 2>/dev/null | grep "LISTEN" | tr -s ' ' | cut -d' ' -f4 | grep -E ":$db_port$" > /dev/null; then
            break
        fi
        sleep 1
    done
    ;;

    sqlite|*)
    rm -f wakapi_testing.db

    echo "Creating database and schema ..."
    sqlite3 wakapi_testing.db < schema.sql

    echo "Importing seed data ..."
    sqlite3 wakapi_testing.db < data.sql

    config="config.sqlite.yml"
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
    sleep 1
    printf "\n"
}

# Run tests
echo "Running Wakapi testing instance in background ..."
echo "Configuration file: $config"
"$initial_run_exe" -config "$config" &
pid=$!
wait_for_wakapi

if [ "$1" == "sqlite" ]; then
    echo "Running test collection ..."
    newman run "wakapi_api_tests.postman_collection.json"
    exit_code=$?
else
    exit_code=0
fi

echo "Shutting down Wakapi ..."
kill -TERM $pid

# Run upgrade tests
if [[ $MIGRATION -eq 1 ]]; then
    echo "Running migrations with build"
    ../wakapi -config "$config" &
    pid=$!

    wait_for_wakapi
    echo "Shutting down Wakapi ..."
    kill -TERM $pid
fi

echo "Exiting with status $exit_code"
exit $exit_code
