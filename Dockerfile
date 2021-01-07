# Build Stage

FROM golang:1.15 AS build-env
WORKDIR /src
ADD ./go.mod .
RUN go mod download

ADD . .
RUN go build -o wakapi

# Run Stage

# When running the application using `docker run`, you can pass environment variables
# to override config values using `-e` syntax.
# Available options can be found in [README.md#-configuration](README.md#-configuration)

FROM debian
WORKDIR /app

ENV ENVIRONMENT prod
ENV WAKAPI_DB_TYPE sqlite3
ENV WAKAPI_DB_USER ''
ENV WAKAPI_DB_PASSWORD ''
ENV WAKAPI_DB_HOST ''
ENV WAKAPI_DB_NAME=/data/wakapi.db
ENV WAKAPI_PASSWORD_SALT ''
ENV WAKAPI_LISTEN_IPV4 '0.0.0.0'
ENV WAKAPI_INSECURE_COOKIES 'true'

COPY --from=build-env /src/wakapi /app/
COPY --from=build-env /src/config.default.yml /app/config.yml
COPY --from=build-env /src/version.txt /app/

RUN sed -i 's/listen_ipv6: ::1/listen_ipv6: /g' /app/config.yml

ADD static /app/static
ADD data /app/data
ADD migrations /app/migrations
ADD views /app/views
ADD wait-for-it.sh .

VOLUME /data

ENTRYPOINT ./wait-for-it.sh
