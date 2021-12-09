#!/bin/sh
###############################################################################################
##               *** WARNING - INSECURE - DO NOT USE IN PRODUCTION ***                       ##
## This script is to simulate operations a trusted orchestrator would perform and as such    ##
## is not a representation of best practices in production environments.                    ##
## https://learn.hashicorp.com/tutorials/vault/pattern-approle?in=vault/recommended-patterns ##
###############################################################################################

# give vault time to come online
sleep 15

# using the token acquire the RoleID for our AppRole
# * typically the RoleID would be embedded in the image and NOT delivered by the trusted orchestrator
# * this is against best practices and included only to simplify the demo
# ref: https://www.vaultproject.io/api-docs/auth/approle#read-approle-role-id
# ref: https://learn.hashicorp.com/tutorials/vault/pattern-approle?in=vault/recommended-patterns
curl --silent \
    --header "X-Vault-Token: ${ORCHESTRATOR_TOKEN}" \
    http://vault:8200/v1/auth/approle/role/dev-role/role-id | jq -r '.data.role_id' > /tmp/role

# using the token generate a new wrapped SecretID on a regular cadence (slightly less than our wrap TTL)
# ref: https://www.vaultproject.io/api-docs/auth/approle#generate-new-secret-id
# ref: https://www.vaultproject.io/docs/concepts/response-wrapping
while true; do
  curl --silent \
      --request POST \
      --header "X-Vault-Token: ${ORCHESTRATOR_TOKEN}" \
      --header "X-Vault-Wrap-TTL: 5m" \
      http://vault:8200/v1/auth/approle/role/dev-role/secret-id | jq -r '.wrap_info.token' > /tmp/secret
    sleep 285
done
