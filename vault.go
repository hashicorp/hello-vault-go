package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	vault "github.com/hashicorp/vault/api"
)

type Vault struct {
	client                  *vault.Client
	role 					string
	auth                    vault.AuthMethod
	databaseCredentialsPath string
	apiKeyPath              string
}

// NewVaultAppRoleClient returns a new client for interacting with Vault KVv2 secrets via AppRole authentication
func NewVaultAppRoleClient(address, approleRoleIDFile, approleSecretIDFile, databaseCredentialsPath, apiKeyPath string) (*Vault, error) {
	config := vault.DefaultConfig() // modify for more granular configuration
	config.Address = address

	client, err := vault.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Vault client: %w", err)
	}

	role, err := ioutil.ReadFile(approleRoleIDFile)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Vault role: %w", err)
	}

	return &Vault{
		client:                  client,
		role:					 string(role),
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
