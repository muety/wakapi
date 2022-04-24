# To build locally: docker buildx build . -t wakapi --load

# Preparation to save some time
FROM --platform=$BUILDPLATFORM golang:1.18-alpine AS prep-env
WORKDIR /src

ADD ./go.mod .
RUN go mod download
ADD . .

RUN wget "https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh" -O wait-for-it.sh && \
    chmod +x wait-for-it.sh

# Build Stage
FROM golang:1.18-alpine AS build-env

# Required for go-sqlite3
RUN apk add --no-cache gcc musl-dev

WORKDIR /src
COPY --from=prep-env /src .

RUN go build -v -o wakapi

WORKDIR /app
RUN cp /src/wakapi . && \
    cp /src/config.default.yml config.yml && \
    sed -i 's/listen_ipv6: ::1/listen_ipv6: /g' config.yml && \
    cp /src/wait-for-it.sh . && \
    cp /src/entrypoint.sh .

# Run Stage

# When running the application using `docker run`, you can pass environment variables
# to override config values using `-e` syntax.
# Available options can be found in [README.md#-configuration](README.md#-configuration)

FROM alpine:3
WORKDIR /app

RUN apk add --no-cache bash ca-certificates tzdata

# See README.md and config.default.yml for all config options
ENV ENVIRONMENT prod
ENV WAKAPI_DB_TYPE sqlite3
ENV WAKAPI_DB_USER ''
ENV WAKAPI_DB_PASSWORD ''
ENV WAKAPI_DB_HOST ''
ENV WAKAPI_DB_NAME=/data/wakapi.db
ENV WAKAPI_PASSWORD_SALT ''
ENV WAKAPI_LISTEN_IPV4 '0.0.0.0'
ENV WAKAPI_INSECURE_COOKIES 'true'
ENV WAKAPI_ALLOW_SIGNUP 'true'

COPY --from=build-env /app .

EXPOSE 3000

ENTRYPOINT /app/entrypoint.sh
