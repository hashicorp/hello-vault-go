#!/bin/sh

docker compose up -d --build

go test

docker compose down --volumes
