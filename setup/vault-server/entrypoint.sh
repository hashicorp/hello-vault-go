#!/bin/sh
export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_SKIP_VERIFY=true
export VAULT_FORMAT='json'

# start our dev server
vault server -dev -dev-listen-address="0.0.0.0:8200" &
sleep 5s

# authenticate containers local vault cli
vault login -no-print "${VAULT_DEV_ROOT_TOKEN_ID}"

# set up vault access for a developer user
vault policy write dev-policy /vault/config/dev-policy.hcl

# enable approle
vault auth enable approle

# create role with some
vault write auth/approle/role/dev-role \
    token_policies=dev-policy \
    secret_id_ttl=1h \
    token_ttl=1h \
    token_max_ttl=30h

# set up kv-v2
vault secrets enable -path=kv-v2 kv-v2

# seed the kv-v2 store with an api key
vault kv put kv-v2/api-key apiKey=my-secret-key

# set up database secrets engine
vault secrets enable database

# configure Vault to be able to connect to DB
vault write database/config/my-postgresql-database \
    plugin_name=postgresql-database-plugin \
    allowed_roles="dev-readonly" \
    connection_url="postgresql://{{username}}:{{password}}@${DB_HOST:="db"}:5432/postgres?sslmode=disable" \
    username="vaultuser" \
    password="vaultpass"

# rotates the password for the Vault user, ensures user is only accessible by Vault itself
vault write -force database/config/my-postgresql-database

# allow Vault to create roles dynamically with the same privileges as the readonly role created in our db init scripts
vault write database/roles/dev-readonly \
    db_name=my-postgresql-database \
    creation_statements="CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}'; \
        GRANT readonly TO \"{{name}}\";" \
    default_ttl="1h" \
    max_ttl="24h"

# write role id to shared location
vault read auth/approle/role/dev-role/role-id | jq -r .data.role_id > /tmp/role

# generate a wrapped secret for the web app to use
# typically this would be handled by a trusted orchestrator
# read more about trusted orchestrators here:
# https://learn.hashicorp.com/tutorials/vault/secure-introduction?in=vault/app-integration#trusted-orchestrator
nohup ./vault/generate-secret.sh &

# run forever
wait