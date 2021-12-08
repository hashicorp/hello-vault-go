package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jessevdk/go-flags"
)

type Environment struct {
	// The address of this service
	MyAddress string `               env:"MY_ADDRESS"                    default:":8080"                        description:"Listen to http traffic on this tcp address"              long:"my-address"`

	// Vault address, approle login credentials, and secret locations
	VaultAddress             string `env:"VAULT_ADDRESS"                 default:"localhost:8200"               description:"Vault address"                                           long:"vault-address"`
	VaultApproleRoleID       string `env:"VAULT_APPROLE_ROLE_ID"         required:"true"                        description:"AppRole role id to authenticate with Vault"              long:"vault-approle-role-id"`
	VaultApproleSecretIDFile string `env:"VAULT_APPROLE_SECRET_ID_FILE"  default:"path/to/wrapping-token"       description:"AppRole secret id file path to authenticate with Vault"  long:"vault-approle-secret-id-file"`
	VaultAPIKeyPath          string `env:"VAULT_API_KEY_PATH"            default:"kv-v2/data/api-key"           description:"Path to the api key used by 'secure-sevice'"             long:"vault-api-key-path"`
	VaultDatabaseCredsPath   string `env:"VAULT_DATABASE_CREDS_PATH"     default:"database/creds/dev-readonly"  description:"Temporary database credentials will be generated here"   long:"vault-database-creds-path"`

	// We will connect to this database using Vault-generated dynamic credentials
	DatabaseHostname string        ` env:"DATABASE_HOSTNAME"             required:"true"                        description:"PostgreSQL database hostname"                            long:"database-hostname"`
	DatabasePort     string        ` env:"DATABASE_PORT"                 default:"5432"                         description:"PostgreSQL database port"                                long:"database-port"`
	DatabaseName     string        ` env:"DATABASE_NAME"                 default:"postgres"                     description:"PostgreSQL database name"                                long:"database-name"`
	DatabaseTimeout  time.Duration ` env:"DATABASE_TIMEOUT"              default:"10s"                          description:"PostgreSQL database connection timeout"                  long:"database-timeout"`

	// A service which requires a specific secret API key (stored in Vault)
	SecureServiceAddress string `    env:"SECURE_SERVICE_ADDRESS"        required:"true"                        description:"3rd party service that requires secure credentials"      long:"secure-service-address"`
}

func main() {
	/* */ log.Println("Hello!")
	defer log.Println("Goodbye!")

	var env Environment

	// parse & validate environment variables
	_, err := flags.Parse(&env)
	if err != nil {
		if flags.WroteHelp(err) {
			os.Exit(0)
		}
		log.Fatalf("Unable to parse environment variables: %v\n", err)
	}

	if err := run(context.Background(), env); err != nil {
		log.Fatalf("Error: %v\n", err)
	}
}

func run(ctx context.Context, env Environment) error {
	// WARNING: the goroutines in this function have simplified error handling
	// and could escape the scope of the function. Production applications
	// may want to add more complex error handling and leak protection logic.

	ctx, cancelContextFunc := context.WithCancel(ctx)
	defer cancelContextFunc()

	// vault
	vault, token, err := NewVaultAppRoleClient(
		ctx,
		VaultParameters{
			env.VaultAddress,
			env.VaultApproleRoleID,
			env.VaultApproleSecretIDFile,
			env.VaultAPIKeyPath,
			env.VaultDatabaseCredsPath,
		},
	)
	if err != nil {
		return fmt.Errorf("unable to initialize vault connection @ %s: %w", env.VaultAddress, err)
	}
	go vault.RenewLoginPeriodically(ctx, token) // keep alive

	// database
	credentials, secret, err := vault.GetDatabaseCredentials(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve database credentials from vault: %w", err)
	}

	database, err := NewDatabase(
		ctx,
		DatabaseParameters{
			hostname: env.DatabaseHostname,
			port:     env.DatabasePort,
			name:     env.DatabaseName,
			timeout:  env.DatabaseTimeout,
		},
		credentials,
	)
	if err != nil {
		return fmt.Errorf("unable to connect to database @ %s:%s: %w", env.DatabaseHostname, env.DatabasePort, err)
	}
	defer func() {
		_ = database.Close()
	}()
	go vault.RenewDatabaseCredentialsPeriodically(ctx, secret, database.Reconnect) // keep alive

	// handlers & routes
	h := Handlers{
		database:             database,
		vault:                vault,
		secureServiceAddress: env.SecureServiceAddress,
	}

	r := gin.Default()

	// demonstrates fetching a static secret from vault and using it to talk to another service
	r.POST("/payments", h.CreatePayment)

	// demonstrates database authentication with dynamic secrets
	r.GET("/products", h.GetProducts)

	r.Run(env.MyAddress)

	return nil
}
