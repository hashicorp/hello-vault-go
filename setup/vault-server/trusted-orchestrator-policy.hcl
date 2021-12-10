# Grant 'update' permission on the 'auth/approle/role/<role_name>/secret-id' path for generating a secret id
path "auth/approle/role/dev-role/secret-id" {
  capabilities = [ "update" ]
}
