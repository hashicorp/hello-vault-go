###############################################################################################
##               *** WARNING - INSECURE - DO NOT USE IN PRODUCTION ***                       ##
## This script is to simulate operations a Vault Operator would perform and as such          ##
## are not a representation of best practices in production environments.                    ##
## https://learn.hashicorp.com/tutorials/vault/pattern-approle?in=vault/recommended-patterns ##
###############################################################################################

# enable a sample TLS listener for our trusted orchestrator to use
# this is a requirement when using cert based authentication
# ref: https://www.vaultproject.io/docs/configuration/listener/tcp#configuring-tls
listener "tcp" {
  address       = "0.0.0.0:8400"
  tls_cert_file = "/vault/certificate.crt"
  tls_key_file  = "/vault/private.pem"
}

ui = true
api_addr = "http://127.0.0.1:8200"