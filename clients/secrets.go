package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	vault "github.com/hashicorp/vault/api"

	"github.com/hashicorp/hello-vault-go/env"
)

type SecretsClient struct {
	Client *vault.Client
}

type DatabaseCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// MustGetSecretsClient returns a new client for interacting with Vault or calls log.Fatal()
func MustGetSecretsClient() *SecretsClient {
	vc, err := NewVaultClient()
	if err != nil {
		log.Fatal("could not get secret store", err)
	}

	sc := &SecretsClient{Client: vc}

	return sc
}

// NewVaultClient returns a new client for interacting with Vault secrets
func NewVaultClient() (*vault.Client, error) {
	config := vault.DefaultConfig() // modify for more granular configuration
	// update address
	config.Address = env.GetOrDefault(env.VaultAddress, "http://localhost:8200")
	client, err := vault.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Vault client: %w", err)
	}

	return client, nil
}

// GetSecret fetches the latest version of a key-value secret (kv-v2)
func (sc SecretsClient) GetSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	secret, err := sc.Client.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read secret: %w", err)
	}

	log.Println("get secret")

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("malformed secret returned")
	}

	return data, nil
}

// PutSecret creates or overwrites a key-value secret (kv-v2)
func (sc SecretsClient) PutSecret(ctx context.Context, path string, data map[string]interface{}) error {
	data = map[string]interface{}{"data": data}

	_, werr := sc.Client.Logical().Write(path, data)
	if werr != nil {
		return fmt.Errorf("unable to write secret: %w", werr)
	}

	log.Println("put secret")
	return nil
}

// GetDatabaseCredentials retrieves a new set of temporary database credentials from Vault
func (sc SecretsClient) GetDatabaseCredentials(ctx context.Context) (*DatabaseCredentials, error) {
	path := env.GetOrDefault(env.VaultDBCredsPath, "database/creds/dev-readonly")
	secret, err := sc.Client.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read secret: %w", err)
	}

	log.Println("get temporary database credentials")

	credsBytes, err := json.Marshal(secret.Data)
	if err != nil {
		return nil, fmt.Errorf("malformed creds returned: %w", err)
	}

	creds := &DatabaseCredentials{}
	uerr := json.Unmarshal(credsBytes, creds)
	if uerr != nil {
		return nil, fmt.Errorf("unable to unmarshal creds: %w", uerr)
	}

	if creds == nil {
		return nil, fmt.Errorf("no database credentials were returned by Vault")
	}

	return creds, nil
}
