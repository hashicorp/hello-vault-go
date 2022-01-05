module github.com/hashicorp/hello-vault-go

go 1.16

require (
	github.com/gin-gonic/gin v1.7.7
	github.com/hashicorp/go-hclog v1.0.0
	github.com/hashicorp/vault v1.9.2
	github.com/hashicorp/vault-plugin-secrets-kv v0.10.1
	github.com/hashicorp/vault/api v1.3.1
	github.com/hashicorp/vault/api/auth/approle v0.1.0
	github.com/hashicorp/vault/sdk v0.3.1-0.20211209192327-a0822e64eae0
	github.com/jessevdk/go-flags v1.5.0
	github.com/lib/pq v1.10.4

)

replace (
github.com/hashicorp/vault/api/auth/approle => github.com/hashicorp/vault/api/auth/approle v0.1.1
github.com/hashicorp/vault/api/auth/userpass => github.com/hashicorp/vault/api/auth/userpass v0.1.0
)