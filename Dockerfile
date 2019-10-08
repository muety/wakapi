# Build Stage
FROM golang:alpine AS build-env
ADD . /src
RUN cd /src && go build -o wakapi

# Final Stage
# When running the application using `docker run`, you can pass environment variables
# to override config values from .env using `-e` syntax.
# Available options are: 
# – WAKAPI_DB_USER
# – WAKAPI_DB_PASSWORD
# – WAKAPI_DB_HOST
# – WAKAPI_DB_NAME
FROM alpine
WORKDIR /app
COPY --from=build-env /src/wakapi /app/
COPY --from=build-env /src/config.ini /app/
COPY --from=build-env /src/.env.example /app/.env
RUN sed -i 's/listen = 127.0.0.1/listen = 0.0.0.0/g' /app/config.ini
ADD static /app/static
ADD wait-for-it.sh .
ENTRYPOINT ./wait-for-it.sh