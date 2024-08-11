FROM golang:alpine AS build-env
WORKDIR /src

RUN wget "https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh" -O wait-for-it.sh && \
    chmod +x wait-for-it.sh

COPY ./go.mod ./go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 go build -ldflags "-s -w" -v -o wakapi main.go

WORKDIR /staging
RUN mkdir ./data ./app && \
    cp /src/wakapi app/ && \
    cp /src/config.default.yml app/config.yml && \
    sed -i 's/listen_ipv6: ::1/listen_ipv6: "-"/g' app/config.yml && \
    cp /src/wait-for-it.sh app/ && \
    cp /src/entrypoint.sh app/ && \
    chown 1000:1000 ./data

# Run Stage

# When running the application using `docker run`, you can pass environment variables
# to override config values using `-e` syntax.
# Available options can be found in [README.md#-configuration](README.md#-configuration)

FROM alpine:3
WORKDIR /app

RUN addgroup -g 1000 app && \
    adduser -u 1000 -G app -s /bin/sh -D app && \
    apk add --no-cache bash ca-certificates tzdata

# See README.md and config.default.yml for all config options
ENV ENVIRONMENT=prod \
    WAKAPI_DB_TYPE=sqlite3 \
    WAKAPI_DB_USER='' \
    WAKAPI_DB_PASSWORD='' \
    WAKAPI_DB_HOST='' \
    WAKAPI_DB_NAME=/data/wakapi.db \
    WAKAPI_PASSWORD_SALT='' \
    WAKAPI_LISTEN_IPV4='0.0.0.0' \
    WAKAPI_INSECURE_COOKIES='true' \
    WAKAPI_ALLOW_SIGNUP='true'

COPY --from=build-env /staging /

LABEL org.opencontainers.image.url="https://github.com/muety/wakapi" \
      org.opencontainers.image.documentation="https://github.com/muety/wakapi" \
      org.opencontainers.image.source="https://github.com/muety/wakapi" \
      org.opencontainers.image.title="Wakapi" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.description="A minimalist, self-hosted WakaTime-compatible backend for coding statistics"

USER app

EXPOSE 3000

ENTRYPOINT /app/entrypoint.sh
