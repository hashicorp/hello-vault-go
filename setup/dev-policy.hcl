# This section grants all access on "secret/*". Further restrictions can be
# applied to this broad policy, as shown below.
path "secret/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

# Even though we allowed secret/*, this line explicitly denies
# secret/super-secret. This takes precedence.
path "secret/super-secret" {
  capabilities = ["deny"]
}

# Allows read-only access to the secret path that will be used
# by Vault to handle generation of dynamic database credentials.
path "database/creds" {
  capabilities = ["read"]
}