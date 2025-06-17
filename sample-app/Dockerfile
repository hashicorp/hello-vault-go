# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

FROM golang:1.23-alpine AS build
COPY  . /app
WORKDIR /app
RUN go build -o hello-vault

FROM alpine:latest
COPY --from=build /app/hello-vault .
EXPOSE 8080
ENTRYPOINT [ "./hello-vault" ]

RUN apk add --no-cache bash curl

HEALTHCHECK \
    --start-period=1s \
    --interval=1s \
    --timeout=1s \
    --retries=30 \
        CMD curl --fail -s http://localhost:8080/healthcheck || exit 1
