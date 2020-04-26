# Build Stage

FROM golang:1.13 AS build-env
ADD . /src
RUN cd /src && go build -o wakapi


# Final Stage

# When running the application using `docker run`, you can pass environment variables
# to override config values from .env using `-e` syntax.
# Available options are: 
# – WAKAPI_DB_TYPE
# – WAKAPI_DB_USER
# – WAKAPI_DB_PASSWORD
# – WAKAPI_DB_HOST
# – WAKAPI_DB_PORT
# – WAKAPI_DB_NAME
# – WAKAPI_DEFAULT_USER_NAME
# – WAKAPI_DEFAULT_USER_PASSWORD

FROM debian
WORKDIR /app

ENV ENV prod
ENV WAKAPI_DB_TYPE sqlite3
ENV WAKAPI_DB_USER ''
ENV WAKAPI_DB_PASSWORD ''
ENV WAKAPI_DB_HOST ''
ENV WAKAPI_DB_NAME=/data/wakapi.db
ENV WAKAPI_DEFAULT_USER_NAME admin
ENV WAKAPI_DEFAULT_USER_PASSWORD admin

COPY --from=build-env /src/wakapi /app/
COPY --from=build-env /src/config.ini /app/
COPY --from=build-env /src/.env.example /app/.env

RUN sed -i 's/listen = 127.0.0.1/listen = 0.0.0.0/g' /app/config.ini

ADD static /app/static
ADD data /app/data
ADD migrations /app/migrations
ADD views /app/views
ADD wait-for-it.sh .

VOLUME /data

ENTRYPOINT ./wait-for-it.sh