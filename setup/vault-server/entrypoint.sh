#!/bin/sh
###############################################################################################
##               *** WARNING - INSECURE - DO NOT USE IN PRODUCTION ***                       ##
## This script is to simulate operations a Vault Operator would perform and as such          ##
## is not a representation of best practices in production environments.                    ##
## https://learn.hashicorp.com/tutorials/vault/pattern-approle?in=vault/recommended-patterns ##
###############################################################################################

export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_FORMAT='json'

# spawn a new process for the development vault server
# and wait for it to come online
# ref: https://www.vaultproject.io/docs/concepts/dev-server
vault server -dev -dev-listen-address="0.0.0.0:8200" &
sleep 5s

# authenticate containers local vault cli
# ref: https://www.vaultproject.io/docs/commands/login
vault login -no-print "${VAULT_DEV_ROOT_TOKEN_ID}"

# Access Policies
# add policies for the various roles we'll be using
# ref: https://www.vaultproject.io/docs/concepts/policies
vault policy write trusted-orchestrator-policy /vault/config/trusted-orchestrator-policy.hcl
vault policy write dev-policy /vault/config/dev-policy.hcl

# AppRole Auth Method
# enable AppRole auth method utilized by our web application
# ref: https://www.vaultproject.io/docs/auth/approle
vault auth enable approle
# configure a specific AppRole role with associated parameters
# ref: https://www.vaultproject.io/api/auth/approle#parameters
# NOTE: token_ttl & token_max_ttl are set to artificially low values to
#       demonstrate the token renewal logic
vault write auth/approle/role/dev-role \
    token_policies=dev-policy \
    secret_id_ttl=1h \
    token_ttl=1m \
    token_max_ttl=3m
# overwrite our RoleID with a known value to simplify our demo
vault write auth/approle/role/dev-role/role-id role_id="${APPROLE_ROLE_ID}"

# Token Auth Method
# configure a token with permissions to act as an orchestrator
# the orchestrator's token will expire after 2 hours
# for brevity we have not introduced a method to renew this token
# ref: https://www.vaultproject.io/docs/commands/token/create
vault token create \
    -id="${ORCHESTRATOR_TOKEN}" \
    -policy=trusted-orchestrator-policy \
    -ttl=2h

# Static Secrets
# enable a kv-v2 secrets engine passing in the path parameter
# ref: https://www.vaultproject.io/docs/secrets/kv/kv-v2
vault secrets enable -path=kv-v2 kv-v2
# seed the kv-v2 store with an entry our web app will use
vault kv put kv-v2/api-key apiKey=my-secret-key

# Dynamic Secrets
# enable a database secrets engine
# ref: https://www.vaultproject.io/docs/secrets/databases
vault secrets enable database
# configure Vault's connection to our db, in this case PostgreSQL
# ref: https://www.vaultproject.io/api/secret/databases/postgresql
vault write database/config/my-postgresql-database \
    plugin_name=postgresql-database-plugin \
    allowed_roles="dev-readonly" \
    connection_url="postgresql://{{username}}:{{password}}@${DB_HOST:="db"}:5432/postgres?sslmode=disable" \
    username="vaultdbuser" \
    password="vaultdbpassword"

# rotate the password for the Vault user, ensures user is only accessible by Vault itself
vault write -force database/config/my-postgresql-database
# allow Vault to create roles dynamically with the same privileges as the
# readonly role created in our database init scripts
# NOTE: default_ttl and max_ttl are set to artificially low values here to
#       demonstrate the credential renewal logic
vault write database/roles/dev-readonly \
    db_name=my-postgresql-database \
    creation_statements="CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}'; \
        GRANT readonly TO \"{{name}}\";" \
    default_ttl="40s" \
    max_ttl="2m"

# keep container alive
tail -f /dev/null & trap 'kill %1' SIGTERM ; wait
