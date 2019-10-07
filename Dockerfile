FROM golang:alpine AS build-env
ADD . /src
RUN cd /src && go build -o wakapi

# final stage
FROM alpine
WORKDIR /app
COPY --from=build-env /src/wakapi /app/
COPY --from=build-env /src/config.ini /app/
COPY --from=build-env /src/.env.example /app/.env
RUN sed -i 's/listen = 127.0.0.1/listen = 0.0.0.0/g' /app/config.ini
ADD static /app/static
ADD wait-for-it.sh .
ENTRYPOINT ./wait-for-it.sh