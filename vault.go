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

	// the locations / field names of our two secrets
	apiKeyPath              string
	apiKeyField             string
	databaseCredentialsPath string
}

type Vault struct {
	client     *vault.Client
	parameters VaultParameters
}

// NewVaultAppRoleClient logs in to Vault using the AppRole authentication
// method, returning an authenticated client and the auth token itself, which
// can be periodically renewed.
func NewVaultAppRoleClient(ctx context.Context, parameters VaultParameters) (*Vault, *vault.Secret, error) {
	log.Printf("connecting to vault @ %s", parameters.address)

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

	log.Println("connecting to vault: success!")

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
	log.Printf("logging in to vault with approle auth; role id: %s", v.parameters.approleRoleID)

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

	authInfo, err := v.client.Auth().Login(ctx, appRoleAuth)
	if err != nil {
		return nil, fmt.Errorf("unable to login using approle auth method: %w", err)
	}
	if authInfo == nil {
		return nil, fmt.Errorf("no approle info was returned after login")
	}

	log.Println("logging in to vault with approle auth: success!")

	return authInfo, nil
}

// GetSecretAPIKey fetches the latest version of secret api key from kv-v2
func (v *Vault) GetSecretAPIKey(ctx context.Context) (string, error) {
	log.Println("getting secret api key from vault")

	secret, err := v.client.Logical().Read(v.parameters.apiKeyPath)
	if err != nil {
		return "", fmt.Errorf("unable to read secret: %w", err)
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("malformed secret returned")
	}

	apiKey, ok := data[v.parameters.apiKeyField]
	if !ok {
		return "", fmt.Errorf("the secret retrieved from vault is missing %q field", v.parameters.apiKeyField)
	}

	apiKeyString, ok := apiKey.(string)
	if !ok {
		return "", fmt.Errorf("unexpected secret key type for %q field", v.parameters.apiKeyField)
	}

	log.Println("getting secret api key from vault: success!")

	return apiKeyString, nil
}

// GetDatabaseCredentials retrieves a new set of temporary database credentials
func (v *Vault) GetDatabaseCredentials(ctx context.Context) (DatabaseCredentials, *vault.Secret, error) {
	log.Println("getting temporary database credentials from vault")

	secret, err := v.client.Logical().Read(v.parameters.databaseCredentialsPath)
	if err != nil {
		return DatabaseCredentials{}, nil, fmt.Errorf("unable to read secret: %w", err)
	}

	credentialsBytes, err := json.Marshal(secret.Data)
	if err != nil {
		return DatabaseCredentials{}, nil, fmt.Errorf("malformed credentials returned: %w", err)
	}

	var credentials DatabaseCredentials

	if err := json.Unmarshal(credentialsBytes, &credentials); err != nil {
		return DatabaseCredentials{}, nil, fmt.Errorf("unable to unmarshal credentials: %w", err)
	}

	log.Println("getting temporary database credentials from vault: success!")

	// raw secret is included to renew database credentials
	return credentials, secret, nil
}
