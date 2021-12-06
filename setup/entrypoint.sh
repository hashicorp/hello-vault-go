#!/bin/sh

set -euo pipefail

## WARNING: It is insecure to configure Vault inside an application's entrypoint script like this.
## An operator should be performing these steps separately. 
## It is only done like this here to simplify the out-of-the-box hello-world experience for the user.

export VAULT_ADDR="http://vault:8200"
export SETUP_TOKEN="root" # WARNING: insecure

# configure Vault server
/app/setup/vault-server/setup.sh

# retrieve role ID for AppRole authentication
export VAULT_APPROLE_ROLE_ID=$(curl -H "X-Vault-Token: ${SETUP_TOKEN}" ${VAULT_ADDR}/v1/auth/approle/role/dev-role/role-id | jq -r .data.role_id)

# start application
/app/bin/hello-vault