# Build Stage

FROM golang:1.15 AS build-env
WORKDIR /src

ADD ./go.mod .
RUN go mod download && go get github.com/markbates/pkger/cmd/pkger

ADD . .
RUN go generate && go build -o wakapi

WORKDIR /app
RUN cp /src/wakapi . && \
    cp /src/config.default.yml config.yml && \
    sed -i 's/listen_ipv6: ::1/listen_ipv6: /g' config.yml && \
    cp /src/wait-for-it.sh .

# Run Stage

# When running the application using `docker run`, you can pass environment variables
# to override config values using `-e` syntax.
# Available options can be found in [README.md#-configuration](README.md#-configuration)

FROM debian
WORKDIR /app

RUN apt update && \
    apt install -y ca-certificates

ENV ENVIRONMENT prod
ENV WAKAPI_DB_TYPE sqlite3
ENV WAKAPI_DB_USER ''
ENV WAKAPI_DB_PASSWORD ''
ENV WAKAPI_DB_HOST ''
ENV WAKAPI_DB_NAME=/data/wakapi.db
ENV WAKAPI_PASSWORD_SALT ''
ENV WAKAPI_LISTEN_IPV4 '0.0.0.0'
ENV WAKAPI_INSECURE_COOKIES 'true'

COPY --from=build-env /app .

VOLUME /data

ENTRYPOINT ./wait-for-it.sh
