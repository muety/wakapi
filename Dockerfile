# Build Stage

FROM golang:1.13 AS build-env
WORKDIR /src
ADD ./go.mod .
RUN go mod download

ADD . .
RUN go build -o wakapi

# Final Stage

# When running the application using `docker run`, you can pass environment variables
# to override config values using `-e` syntax.
# Available options are:
# – WAKAPI_DB_TYPE
# – WAKAPI_DB_USER
# – WAKAPI_DB_PASSWORD
# – WAKAPI_DB_HOST
# – WAKAPI_DB_PORT
# – WAKAPI_DB_NAME
# – WAKAPI_PASSWORD_SALT
# – WAKAPI_BASE_PATH

FROM debian
WORKDIR /app

ENV ENVIRONMENT prod
ENV WAKAPI_DB_TYPE sqlite3
ENV WAKAPI_DB_USER ''
ENV WAKAPI_DB_PASSWORD ''
ENV WAKAPI_DB_HOST ''
ENV WAKAPI_DB_NAME=/data/wakapi.db
ENV WAKAPI_PASSWORD_SALT ''

COPY --from=build-env /src/wakapi /app/
COPY --from=build-env /src/config.default.yml /app/config.yml
COPY --from=build-env /src/version.txt /app/

RUN sed -i 's/listen_ipv4: 127.0.0.1/listen_ipv4: 0.0.0.0/g' /app/config.yml
RUN sed -i 's/insecure_cookies: false/insecure_cookies: true/g' /app/config.yml

ADD static /app/static
ADD data /app/data
ADD migrations /app/migrations
ADD views /app/views
ADD wait-for-it.sh .

VOLUME /data

ENTRYPOINT ./wait-for-it.sh
