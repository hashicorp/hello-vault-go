#!/bin/sh

# WARNING: THIS SCRIPT IS NOT A REFLECTION OF SECURITY BEST PRACTICES.
# DO NOT USE IN PRODUCTION.

# This script is a simplification of what the "trusted orchestrator" should do to generate 
# the secret ID required for the application to authenticate to Vault.
#
# Realistically, your trusted orchestrator should be something like Kubernetes, Chef, Terraform--whatever it is 
# that you trust to launch your application. 
# This curl script is used here merely to simplify this demo application by reducing dependencies.

# Read more about trusted orchestrators here: 
# https://learn.hashicorp.com/tutorials/vault/secure-introduction?in=vault/app-integration#trusted-orchestrator

mkdir -p /app/path/to

while true; do
    curl -X PUT -H "X-Vault-Token: ${VAULT_TOKEN}" -H "X-Vault-Wrap-Ttl: 5m10s" \
    -d "null" ${VAULT_ADDR}/v1/auth/approle/role/dev-role/secret-id | jq -r .wrap_info.token > /app/path/to/wrapping-token

    sleep 5m
done
