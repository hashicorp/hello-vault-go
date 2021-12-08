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

# create role with policies and ttls
# - secret_id_ttl:   determines how long the unwrapped secret_id will be valid for
# - token_ttl:       determines how long the login token is valid
# - token_max_ttl:   determines how long the login token can be renewed
# read more here about parameters here:
# https://www.vaultproject.io/api/auth/approle#parameters
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

# Typically these steps (provisioning a role and wrapped secret as well as making it available to the container)
# are handled by a trusted orchestrator. In this example we're using a docker shared volume as our trusted entity
# read more about trusted orchestrators here:
# https://learn.hashicorp.com/tutorials/vault/secure-introduction?in=vault/app-integration#trusted-orchestrator
# write role id to shared location for the web app to consume
vault read auth/approle/role/dev-role/role-id | jq -r .data.role_id > /tmp/role
# generate a wrapped secret (one time use) that expires after 10 seconds for the web app to consume
vault write -wrap-ttl=10s -f auth/approle/role/dev-role/secret-id | jq -r .wrap_info.token > /tmp/secret

# keep vault container alive
wait