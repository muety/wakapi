#!/bin/bash

# usage: file_env VAR [DEFAULT]
#    ie: file_env 'XYZ_DB_PASSWORD' 'example'
# (will allow for "$XYZ_DB_PASSWORD_FILE" to fill in the value of
#  "$XYZ_DB_PASSWORD" from a file, especially for Docker's secrets feature)
file_env() {
	local var="$1"
	local fileVar="${var}_FILE"
	local def="${2:-}"
	if [ "${!var:-}" ] && [ "${!fileVar:-}" ]; then
		printf >&2 'error: both %s and %s are set (but are exclusive)\n' "$var" "$fileVar"
		exit 1
	fi
	local val="$def"
	if [ "${!var:-}" ]; then
		val="${!var}"
	elif [ "${!fileVar:-}" ]; then
		val="$(< "${!fileVar}")"
	fi
	export "$var"="$val"
	unset "$fileVar"
}

file_env "WAKAPI_PASSWORD_SALT"
file_env "WAKAPI_DB_PASSWORD"
file_env "WAKAPI_MAIL_SMTP_PASS"
file_env "WAKAPI_SUBSCRIPTIONS_STRIPE_SECRET_KEY"
file_env "WAKAPI_SUBSCRIPTIONS_STRIPE_ENDPOINT_SECRET"

if [ "$WAKAPI_DB_TYPE" == "sqlite3" ] || [ "$WAKAPI_DB_TYPE" == "" ]; then
  exec ./wakapi
else
  echo "Waiting for database to come up"
  exec ./wait-for-it.sh "$WAKAPI_DB_HOST:$WAKAPI_DB_PORT" -s -t 60 -- ./wakapi
fi
