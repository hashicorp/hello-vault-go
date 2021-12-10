#!/bin/sh
###############################################################################################
##               *** WARNING - INSECURE - DO NOT USE IN PRODUCTION ***                       ##
## This script is to simulate operations a Vault Operator would perform and as such          ##
## are not a representation of best practices in production environments.                    ##
## https://learn.hashicorp.com/tutorials/vault/pattern-approle?in=vault/recommended-patterns ##
###############################################################################################

# give vault time to come online
sleep 15

# retrieve a vault token for our trusted orchestrator using a trusted self signed certificate
# ref: https://www.vaultproject.io/api/auth/cert#login-with-tls-certificate-method
login_token=$(curl --silent \
   --request POST \
   --key /certs/private.pem \
   --cert /certs/certificate.crt \
   --insecure \
   -H "Content-Type: application/json" \
   -d '{"name": "trusted-orchestrator"}' \
   https://vault:8400/v1/auth/cert/login | jq -r '.auth.client_token')

# using the retrieved token acquire the RoleID for our AppRole
# * typically the RoleID would be embedded in the image and NOT delivered by the trusted orchestrator
# * this is against best practices and included only to simplify the demo
# ref: https://www.vaultproject.io/api-docs/auth/approle#read-approle-role-id
# ref: https://learn.hashicorp.com/tutorials/vault/pattern-approle?in=vault/recommended-patterns
curl --silent \
    --header "X-Vault-Token: ${login_token}" \
    --insecure \
    https://vault:8400/v1/auth/approle/role/dev-role/role-id | jq -r '.data.role_id' > /tmp/role

# using the retrieved login_token generate a new wrapped SecretID
# ref: https://www.vaultproject.io/api-docs/auth/approle#generate-new-secret-id
# ref: https://www.vaultproject.io/docs/concepts/response-wrapping
curl --silent \
    --request POST \
    --header "X-Vault-Token: ${login_token}" \
    --header "X-Vault-Wrap-TTL: 5m" \
    --insecure \
    https://vault:8400/v1/auth/approle/role/dev-role/secret-id | jq -r '.wrap_info.token' > /tmp/secret