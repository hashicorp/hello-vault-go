#!/bin/sh

###############################################################################################
##               *** WARNING - INSECURE - DO NOT USE IN PRODUCTION ***                       ##
## This script is to simulate operations a Vault operator would perform and, as such,        ##
## is not a representation of best practices in production environments.                     ##
## https://learn.hashicorp.com/tutorials/vault/pattern-approle?in=vault/recommended-patterns ##
###############################################################################################

set -e

export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_FORMAT='json'

# spawn a new process for the development vault server and wait for it to come online
# ref: https://www.vaultproject.io/docs/concepts/dev-server
vault server -dev -dev-listen-address="0.0.0.0:8200" &
sleep 5s

# authenticate container's local vault cli
# ref: https://www.vaultproject.io/docs/commands/login
vault login -no-print "${VAULT_DEV_ROOT_TOKEN_ID}"

#####################################
########## ACCESS POLICIES ##########
#####################################

# add policies for the various roles we'll be using
# ref: https://www.vaultproject.io/docs/concepts/policies
vault policy write trusted-orchestrator-policy /vault/config/trusted-orchestrator-policy.hcl
vault policy write dev-policy /vault/config/dev-policy.hcl

#####################################
######## APPROLE AUTH METHDO ########
#####################################

# enable AppRole auth method utilized by our web application
# ref: https://www.vaultproject.io/docs/auth/approle
vault auth enable approle

# configure a specific AppRole role with associated parameters
# ref: https://www.vaultproject.io/api/auth/approle#parameters
vault write auth/approle/role/dev-role \
    token_policies=dev-policy \
    secret_id_ttl="2m" \
    token_ttl="2m" \
    token_max_ttl="6m"  # artificially low ttl to demonstrate token renewal

# overwrite our RoleID with a known value to simplify our demo
vault write auth/approle/role/dev-role/role-id role_id="${APPROLE_ROLE_ID}"

#####################################
######### TOKEN AUTH METHOD #########
#####################################

# configure a token with permissions to act as a trusted orchestrator
# nb: for simplicity, we don't handle renewals in our simulated orchestrator
# so we've set the ttl to a very long duration (768h); when this expires
# the web app will no longer receive a SecretID and subsequently fail on the
# next attempted AppRole login
# ref: https://www.vaultproject.io/docs/commands/token/create
vault token create \
    -id="${ORCHESTRATOR_TOKEN}" \
    -policy=trusted-orchestrator-policy \
    -ttl="768h"

#####################################
########## STATIC SECRETS ###########
#####################################

# enable a kv-v2 secrets engine, passing in the path parameter
# ref: https://www.vaultproject.io/docs/secrets/kv/kv-v2
vault secrets enable -path=kv-v2 kv-v2

# seed the kv-v2 store with an entry our web app will use
vault kv put "${API_KEY_PATH}" "${API_KEY_FIELD}=my-secret-key"

#####################################
########## DYNAMIC SECRETS ##########
#####################################

# enable a database secrets engine
# ref: https://www.vaultproject.io/docs/secrets/databases
vault secrets enable database

# configure Vault's connection to our db, in this case PostgreSQL
# ref: https://www.vaultproject.io/api/secret/databases/postgresql
vault write database/config/my-postgresql-database \
    plugin_name=postgresql-database-plugin \
    allowed_roles="dev-readonly" \
    connection_url="postgresql://{{username}}:{{password}}@${DATABASE_HOSTNAME}:${DATABASE_PORT}/postgres?sslmode=disable" \
    username="vault_db_user" \
    password="vault_db_password"

# rotate the password for 'vault_db_user', ensures the user is only accessible by Vault itself
vault write -force database/config/my-postgresql-database

# Allow vault to create roles dynamically with the same privileges as the "readonly"
# role created in our database's 3-roles.sql script.
#
# NOTE: we use artificially low ttl values to demonstrate the credential renewal logic
vault write database/roles/dev-readonly \
    db_name=my-postgresql-database \
    creation_statements="CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}'; GRANT readonly TO \"{{name}}\";" \
    renew_statements="ALTER ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}'; GRANT readonly TO \"{{name}}\";" \
    default_ttl="1m" \
    max_ttl="3m"

# this container is now healthy
touch /tmp/healthy

# keep container alive
tail -f /dev/null & trap 'kill %1' TERM ; wait
