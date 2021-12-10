#!/bin/sh
###############################################################################################
##               *** WARNING - INSECURE - DO NOT USE IN PRODUCTION ***                       ##
## This script is to simulate operations a trusted orchestrator would perform and as such    ##
## is not a representation of best practices in production environments.                     ##
## Normally the trusted orchestrator is the mechanism that launches applications and injects ##
## them with a Secret ID at runtime; typically something like Terraform, K8s, or Chef.       ##
## https://learn.hashicorp.com/tutorials/vault/secure-introduction#trusted-orchestrator      ##
###############################################################################################

# give vault time to come online
sleep 15

# using the orchestrator token generate a new wrapped SecretID on a regular
# cadence (slightly less than our wrap TTL)
# ref: https://www.vaultproject.io/api-docs/auth/approle#generate-new-secret-id
# ref: https://www.vaultproject.io/docs/concepts/response-wrapping
while true; do
  curl --silent \
       --request POST \
       --header "X-Vault-Token: ${ORCHESTRATOR_TOKEN}" \
       --header "X-Vault-Wrap-TTL: 5m" \
          http://vault:8200/v1/auth/approle/role/dev-role/secret-id | jq -r '.wrap_info.token' > /tmp/secret

  # sleep for a very short time to demonstrate our token renewal logic
  sleep 30
done
