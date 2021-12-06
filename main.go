package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jessevdk/go-flags"
)

// Environment describes the environment variables (or equivalent command-line
// flags) that will be parsed and populated at startup, the default values will
// be used unless overwritten
type Environment struct {
	// The address of this service
	MyAddress string `               env:"MY_ADDRESS"                    default:":8080"                        description:"Listen to http traffic on this tcp address"              long:"my-address"`

	// Vault address, approle login credentialials to authenticate with Vault,
	// and the paths where secrets are to be found
	VaultAddress             string `env:"VAULT_ADDRESS"                 default:"localhost:8200"               description:"Vault address"                                           long:"vault-address"`
	VaultApproleRoleIDFile   string `env:"VAULT_APPROLE_ROLE_ID_FILE"    default:"/tmp/role"                    description:"AppRole role id file pathto authenticate with Vault"     long:"vault-approle-role-id-file"`
	VaultApproleSecretIDFile string `env:"VAULT_APPROLE_SECRET_ID_FILE"  default:"/tmp/secret"       			description:"AppRole secret id file path to authenticate with Vault"  long:"vault-approle-secret-id-file"`
	VaultDatabaseCredsPath   string `env:"VAULT_DATABASE_CREDS_PATH"     default:"database/creds/dev-readonly"  description:"Temporary database credentials will be generated here"   long:"vault-database-creds-path"`
	VaultAPIKeyPath          string `env:"VAULT_API_KEY_PATH"            default:"kv-v2/data/api-key"           description:"Path to the api key used by 'secure-sevice'"             long:"vault-api-key-path"`

	// We will connect to this database using Vault-generated dynamic credentials
	DatabaseHostname string        ` env:"DATABASE_HOSTNAME"             required:"true"                        description:"PostgreSQL database hostname"                            long:"database-hostname"`
	DatabasePort     string        ` env:"DATABASE_PORT"                 default:"5432"                         description:"PostgreSQL database port"                                long:"database-port"`
	DatabaseName     string        ` env:"DATABASE_NAME"                 default:"postgres"                     description:"PostgreSQL database name"                                long:"database-name"`
	DatabaseTimeout  time.Duration ` env:"DATABASE_TIMEOUT"              default:"10s"                          description:"PostgreSQL database connection timeout"                  long:"database-timeout"`

	// A service which requires a specific secret API key (stored in Vault) to be
	// provided in the request header
	SecureServiceAddress string `    env:"SECURE_SERVICE_ADDRESS"        required:"true"                        description:"3rd party service that requires secure credentials"      long:"secure-service-address"`
}

func main() {
	var env Environment

	// parse & validate environment variables
	_, err := flags.Parse(&env)
	if err != nil {
		if flags.WroteHelp(err) {
			os.Exit(0)
		}
		log.Fatalf("Unable to parse environment variables: %v\n", err)
	}

	// ctx := context.Background()

	// vault
	vault, err := NewVaultAppRoleClient(
		env.VaultAddress,
		env.VaultApproleRoleIDFile,
		env.VaultApproleSecretIDFile,
		env.VaultDatabaseCredsPath,
		env.VaultAPIKeyPath,
	)
	if err != nil {
		log.Fatalf("Unable to initialize vault connection @ %s: %v\n", env.VaultAddress, err)
	}

	// database
	database, err := NewDatabase(
		ctx,
		env.DatabaseHostname,
		env.DatabasePort,
		env.DatabaseName,
		env.DatabaseTimeout,
		vault,
	)
	if err != nil {
		log.Fatalf("Unable to connect to database @ %s:%s: %v\n", env.DatabaseHostname, env.DatabasePort, err)
	}
	defer func() {
		_ = database.Close()
	}()

	// handlers & routes
	h := Handlers{
		//database:             database,
		vault:                vault,
		secureServiceAddress: env.SecureServiceAddress,
	}

	r := gin.Default()

	// demonstrates fetching a static secret from vault and using it to talk to another service
	r.POST("/payments", h.CreatePayment)

	// demonstrates database authentication with dynamic secrets
	r.GET("/products", h.GetProducts)

	r.Run(env.MyAddress)
}
