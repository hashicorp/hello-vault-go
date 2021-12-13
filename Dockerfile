FROM golang:1.16-alpine AS build
COPY  . /app
WORKDIR /app
RUN go build -o hello-vault

FROM alpine:latest
COPY --from=build /app/hello-vault .
EXPOSE 8080
ENTRYPOINT [ "./hello-vault" ]

RUN apk add --no-cache curl

HEALTHCHECK \
    --start-period=1s \
    --interval=1s \
    --timeout=1s \
    --retries=30 \
        CMD curl --fail -s http://localhost:8080/healthcheck || exit 1
