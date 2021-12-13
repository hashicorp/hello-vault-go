# This section grants access to "kv-v2/data/api-key"
path "kv-v2/data/api-key" {
  capabilities = ["read", "update"]
}

# Allows read-only access to the secret path that will be used
# by Vault to handle generation of dynamic database credentials.
path "database/creds/dev-readonly" {
  capabilities = ["read"]
}
