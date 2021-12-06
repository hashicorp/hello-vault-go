#!/bin/sh

# (re)start application, Vault server, and database
docker-compose down
docker-compose build
docker-compose up -d