FROM --platform=$BUILDPLATFORM golang:alpine AS build-env
WORKDIR /src

RUN apk add --no-cache ca-certificates tzdata && update-ca-certificates

COPY ./go.mod ./go.sum ./
RUN go mod download
COPY . .

ARG TARGETOS
ARG TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 GOEXPERIMENT=greenteagc,jsonv2 go build -ldflags "-s -w" -v -o wakapi main.go

WORKDIR /staging
RUN mkdir ./data ./app && \
    cp /src/wakapi app/ && \
    cp /src/config.default.yml app/config.yml && \
    sed -i 's/listen_ipv6: ::1/listen_ipv6: "-"/g' app/config.yml

# Run Stage

# When running the application using `docker run`, you can pass environment variables
# to override config values using `-e` syntax.
# Available options can be found in [README.md#-configuration](README.md#-configuration)

# Note on the distroless image:
# we could use `base:nonroot`, which already includes ca-certificates and tz, but that one it actually larger than alpine,
# probably because of glibc, whereas alpine uses musl. The `static:nonroot`, doesn't include any libc implementation, because only meant for true static binaries without cgo, etc.

FROM gcr.io/distroless/static:nonroot
WORKDIR /app

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

COPY --from=build-env --chown=nonroot:nonroot /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build-env --chown=nonroot:nonroot /usr/share/zoneinfo /usr/share/zoneinfo

COPY --from=build-env --chown=nonroot:nonroot /staging/app /app
COPY --from=build-env --chown=nonroot:nonroot /staging/data /data

LABEL org.opencontainers.image.url="https://github.com/muety/wakapi" \
    org.opencontainers.image.documentation="https://github.com/muety/wakapi" \
    org.opencontainers.image.source="https://github.com/muety/wakapi" \
    org.opencontainers.image.title="Wakapi" \
    org.opencontainers.image.licenses="MIT" \
    org.opencontainers.image.description="A minimalist, self-hosted WakaTime-compatible backend for coding statistics"

USER nonroot

EXPOSE 3000

ENTRYPOINT ["/app/wakapi"]
