package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	vault "github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/approle"
)

type Vault struct {
	client                  *vault.Client
	auth                    vault.AuthMethod
	role					string
	databaseCredentialsPath string
	apiKeyPath              string
}

// NewVaultAppRoleClient returns a new client for interacting with Vault KVv2 secrets via AppRole authentication
func NewVaultAppRoleClient(address, approleRoleID, approleSecretIDFile, databaseCredentialsPath, apiKeyPath string) (*Vault, error) {
	config := vault.DefaultConfig() // modify for more granular configuration
	config.Address = address

	client, err := vault.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Vault client: %w", err)
	}

	// A combination of a Role ID and a Secret ID is required to log into Vault
	// with AppRole authentication method. The Secret ID is a value that needs
	// to be protected, so instead of the app having knowledge of the secret ID
	// directly, we have a trusted orchestrator
	// (https://learn.hashicorp.com/tutorials/vault/secure-introduction?in=vault/app-integration#trusted-orchestrator)
	// give the app accev to a short-lived response-wrapping token
	// (https://www.vaultproject.io/docs/concepts/response-wrapping).
	// Read more at: https://learn.hashicorp.com/tutorials/vault/approle-best-practices?in=vault/auth-methods#secretid-delivery-best-practices
	approleSecretID := &approle.SecretID{
		FromFile: approleSecretIDFile,
	}

	appRoleAuth, err := approle.NewAppRoleAuth(
		approleRoleID,
		approleSecretID,
		approle.WithWrappingToken(), // Only required if the secret ID is response-wrapped.
	)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize approle authentication method: %w", err)
	}

	return &Vault{
		client:                  client,
		role:					 approleRoleID,
		auth:                    appRoleAuth,
		databaseCredentialsPath: databaseCredentialsPath,
		apiKeyPath:              apiKeyPath,
	}, nil
}

// GetSecretAPIKey fetches the latest version of secret api key from kv-v2
func (v *Vault) GetSecretAPIKey(ctx context.Context) (map[string]interface{}, error) {
	secret, err := v.client.Logical().Read(v.apiKeyPath)
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

type DatabaseCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// GetDatabaseCredentials retrieves a new set of temporary database credentials from Vault
func (v *Vault) GetDatabaseCredentials(ctx context.Context) (DatabaseCredentials, error) {
	secret, err := v.client.Logical().Read(v.databaseCredentialsPath)
	if err != nil {
		return DatabaseCredentials{}, fmt.Errorf("unable to read secret: %w", err)
	}

	log.Println("got temporary database credentials")

	credsBytes, err := json.Marshal(secret.Data)
	if err != nil {
		return DatabaseCredentials{}, fmt.Errorf("malformed creds returned: %w", err)
	}

	var credentials DatabaseCredentials

	if err := json.Unmarshal(credsBytes, &credentials); err != nil {
		return DatabaseCredentials{}, fmt.Errorf("unable to unmarshal creds: %w", err)
	}

	return credentials, nil
}

// PutSecret creates or overwrites a key-value secret (kv-v2) after authenticating via AppRole
func (v *Vault) PutSecret(ctx context.Context, path string, data map[string]interface{}) error {
	data = map[string]interface{}{"data": data}

	_, err := v.client.Logical().Write(path, data)
	if err != nil {
		return fmt.Errorf("unable to write secret: %w", err)
	}

	log.Println("put secret")

	return nil
}
