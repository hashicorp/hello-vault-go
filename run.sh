#!/bin/sh

# (re)start application and its dependencies
docker compose down --volumes
docker compose up -d --build
