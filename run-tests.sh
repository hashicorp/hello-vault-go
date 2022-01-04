#!/bin/sh

docker compose up -d --build --quiet-pull

go test -v

docker compose down --volumes
