package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	vault "github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/approle"
)

type VaultParameters struct {
	// connection parameters
	address             string
	approleRoleID       string
	approleSecretIDFile string

	// the locations of our two secrets
	apiKeyPath              string
	databaseCredentialsPath string
}

type Vault struct {
	client     *vault.Client
	parameters VaultParameters
}

// NewVaultAppRoleClient logs in to Vault using AppRole authentication method, returns a client & the auth token
func NewVaultAppRoleClient(ctx context.Context, parameters VaultParameters) (*Vault, *vault.Secret, error) {
	log.Println("Connecting to vault @", parameters.address)

	config := vault.DefaultConfig() // modify for more granular configuration
	config.Address = parameters.address

	client, err := vault.NewClient(config)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to initialize vault client: %w", err)
	}

	vault := &Vault{
		client:     client,
		parameters: parameters,
	}

	token, err := vault.login(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("vault login error: %w", err)
	}

	return vault, token, nil
}

// A combination of a RoleID and a SecretID is required to log into Vault
// with AppRole authentication method. The SecretID is a value that needs
// to be protected, so instead of the app having knowledge of the SecretID
// directly, we have a trusted orchestrator (simulated with a script here)
// give the app access to a short-lived response-wrapping token.
//
// ref: https://www.vaultproject.io/docs/concepts/response-wrapping
// ref: https://learn.hashicorp.com/tutorials/vault/secure-introduction?in=vault/app-integration#trusted-orchestrator
// ref: https://learn.hashicorp.com/tutorials/vault/approle-best-practices?in=vault/auth-methods#secretid-delivery-best-practices
func (v *Vault) login(ctx context.Context) (*vault.Secret, error) {
	approleSecretID := &approle.SecretID{
		FromFile: v.parameters.approleSecretIDFile,
	}

	appRoleAuth, err := approle.NewAppRoleAuth(
		v.parameters.approleRoleID,
		approleSecretID,
		approle.WithWrappingToken(), // only required if the SecretID is response-wrapped
	)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize approle authentication method: %w", err)
	}

	log.Println("Attempting to log in to vault using RoleID", v.parameters.approleRoleID)

	authInfo, err := v.client.Auth().Login(ctx, appRoleAuth)
	if err != nil {
		return nil, fmt.Errorf("unable to login using approle auth method: %w", err)
	}
	if authInfo == nil {
		return nil, fmt.Errorf("no approle info was returned after login")
	}

	log.Println("Successfully logged in to vault")

	return authInfo, nil
}

// GetSecretAPIKey fetches the latest version of secret api key from kv-v2
func (v *Vault) GetSecretAPIKey(ctx context.Context) (map[string]interface{}, error) {
	// path starting with secret/
	secret, err := v.client.Logical().Read(v.parameters.apiKeyPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read secret: %w", err)
	}

	log.Println("Got secret api key")

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("malformed secret returned")
	}

	return data, nil
}

// GetDatabaseCredentials retrieves a new set of temporary database credentials
func (v *Vault) GetDatabaseCredentials(ctx context.Context) (DatabaseCredentials, *vault.Secret, error) {
	// path starting with database/
	secret, err := v.client.Logical().Read(v.parameters.databaseCredentialsPath)
	if err != nil {
		return DatabaseCredentials{}, nil, fmt.Errorf("unable to read secret: %w", err)
	}

	log.Println("Got temporary database credentials")

	credentialsBytes, err := json.Marshal(secret.Data)
	if err != nil {
		return DatabaseCredentials{}, nil, fmt.Errorf("malformed credentials returned: %w", err)
	}

	var credentials DatabaseCredentials

	if err := json.Unmarshal(credentialsBytes, &credentials); err != nil {
		return DatabaseCredentials{}, nil, fmt.Errorf("unable to unmarshal credentials: %w", err)
	}

	// raw secret is included to renew database credentials
	return credentials, secret, nil
}
