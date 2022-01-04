#!/bin/sh

docker compose down --volumes
docker compose up -d --build

go test

docker compose down --volumes
