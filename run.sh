#!/bin/sh

# (re)start application, Vault server, and database
docker-compose down
docker-compose up -d