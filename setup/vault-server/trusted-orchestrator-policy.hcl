###############################################################################################
##               *** WARNING - INSECURE - DO NOT USE IN PRODUCTION ***                       ##
## This script is to simulate operations a Vault Operator would perform and as such          ##
## are not a representation of best practices in production environments.                    ##
## https://learn.hashicorp.com/tutorials/vault/pattern-approle?in=vault/recommended-patterns ##
###############################################################################################

# Grant 'read' permission on the 'auth/approle/role/<role_name>/role-id' path
path "auth/approle/role/dev-role/role-id" {
  capabilities = [ "read" ]
}

# Grant 'update' permission on the 'auth/approle/role/<role_name>/secret-id' path for generating a secret id
path "auth/approle/role/dev-role/secret-id" {
  capabilities = [ "update" ]
}