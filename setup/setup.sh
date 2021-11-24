#!/bin/bash

set -euo pipefail

export VAULT_ADDR="http://0.0.0.0:8200"
export VAULT_TOKEN="root"

# (re)start application, Vault server, and database
cd ..
docker-compose down
docker-compose build
docker-compose up -d
cd setup/

# give Vault a moment to come up fully before pinging it
sleep 1

# set up vault access for a developer user
vault policy write dev-policy dev-policy.hcl

vault auth enable approle

vault write auth/approle/role/dev-role \
    token_policies=dev-policy

# set up database secrets engine
vault secrets enable database

# configure Vault to be able to connect to DB
vault write database/config/my-postgresql-database \
    plugin_name=postgresql-database-plugin \
    allowed_roles="dev-readonly" \
    connection_url="postgresql://{{username}}:{{password}}@db:5432/postgres?sslmode=disable" \
    username="vaultuser" \
    password="vaultpass"

# allow Vault to create roles dynamically with the same privileges as the readonly role created in our db init scripts
vault write database/roles/dev-readonly \
    db_name=my-postgresql-database \
    creation_statements="CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}'; \
        GRANT readonly TO \"{{name}}\";" \
    default_ttl="1h" \
    max_ttl="24h"

