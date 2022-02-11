#!/bin/sh

echo "Starting Vault dev server.."
container_id=$(docker run --rm --detach -p 8200:8200 -e 'VAULT_DEV_ROOT_TOKEN_ID=dev-only-token' vault)

echo "Running quickstart example."
go run main.go

echo "Stopping Vault dev server.."
docker stop "${container_id}" > /dev/null

echo "Vault server has stopped."
