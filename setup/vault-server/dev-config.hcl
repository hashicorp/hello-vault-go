#enable a sample TLS listener for our trusted orchestrator to use
#ref: https://www.vaultproject.io/docs/configuration/listener/tcp#configuring-tls
listener "tcp" {
  address       = "0.0.0.0:8400"
  tls_cert_file = "/vault/certificate.crt"
  tls_key_file  = "/vault/private.pem"
}

ui=true
api_addr="http://127.0.0.1:8200"