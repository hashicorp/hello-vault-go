#!/bin/sh
# WARNING: THIS SCRIPT IS NOT A REFLECTION OF SECURITY BEST PRACTICES.
# DO NOT USE IN PRODUCTION.

# This script is a simplification of what the "trusted orchestrator" should do to generate
# a wrapped secret ID and make it available to the container.
#
# This curl script is used here merely to simplify this demo application by reducing dependencies.
#
# Read more about trusted orchestrators here:
# https://learn.hashicorp.com/tutorials/vault/secure-introduction?in=vault/app-integration#trusted-orchestrator

while true; do
  curl -X PUT -H "X-Vault-Token: ${VAULT_DEV_ROOT_TOKEN_ID}" -H "X-Vault-Wrap-Ttl: 5m" \
  -d "null" "${VAULT_ADDR}"/v1/auth/approle/role/dev-role/secret-id | jq -r .wrap_info.token > /tmp/secret
  sleep 4m 45s
done