#!/bin/bash

set -euo pipefail

export VAULT_ADDR="http://0.0.0.0:8200"
export VAULT_TOKEN="root"

docker-compose down
docker-compose up -d

# give Vault a moment to come up fully before pinging it
sleep 1

# set up vault access for a developer user
vault policy write dev-policy dev-policy.hcl

vault auth enable approle

vault write auth/approle/role/dev-role \
    token_policies=dev-policy