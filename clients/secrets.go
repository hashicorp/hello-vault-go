package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	vault "github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/approle"

	"github.com/hashicorp/hello-vault-go/env"
)

type DatabaseCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type secretsClient struct {
	vc   *vault.Client
	auth vault.AuthMethod
}

// MustGetVaultAppRoleClient returns a new client for interacting with Vault or calls log.Fatal()
func MustGetVaultAppRoleClient() (ss *secretsClient) {
	ss, err := NewVaultAppRoleClient()

	if err != nil {
		log.Fatal("could not get secret store", err)
	}
	return
}

// NewVaultAppRoleClient returns a new client for interacting with Vault KVv2 secrets via AppRole authentication
func NewVaultAppRoleClient() (*secretsClient, error) {
	ss := &secretsClient{}
	config := vault.DefaultConfig() // modify for more granular configuration
	//update address
	config.Address = env.GetOrDefault(env.VaultAddress, "http://localhost:8200")
	client, err := vault.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Vault client: %w", err)
	}

	ss.vc = client

	// A combination of a Role ID and Secret ID is required to log in to Vault
	// with an AppRole. We're passing this in from an environment variable, "APPROLE_ROLE_ID".
	role := env.MustGet(env.AppRoleID)

	// The Secret ID is a value that needs to be protected, so instead of the
	// app having knowledge of the secret ID directly, we have a trusted orchestrator (https://learn.hashicorp.com/tutorials/vault/secure-introduction?in=vault/app-integration#trusted-orchestrator)
	// give the app access to a short-lived response-wrapping token (https://www.vaultproject.io/docs/concepts/response-wrapping).
	// Read more at: https://learn.hashicorp.com/tutorials/vault/approle-best-practices?in=vault/auth-methods#secretid-delivery-best-practices
	secretID := &approle.SecretID{FromFile: "path/to/wrapping-token"}

	appRoleAuth, err := approle.NewAppRoleAuth(
		role,
		secretID,
		approle.WithWrappingToken(), // Only required if the secret ID is response-wrapped.
	)

	if err != nil {
		return nil, fmt.Errorf("unable to initialize AppRole approle method: %w", err)
	}
	ss.auth = appRoleAuth
	return ss, nil
}

// GetSecret fetches the latest version of a key-value secret (kv-v2)
func (ss secretsClient) GetSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	authInfo, err := ss.vc.Auth().Login(ctx, ss.auth)
	if err != nil {
		return nil, fmt.Errorf("unable to login to AppRole approle method: %w", err)
	}
	if authInfo == nil {
		return nil, fmt.Errorf("no approle info was returned after login")
	}

	secret, err := ss.vc.Logical().Read(path)
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

// PutSecret creates or overwrites a key-value secret (kv-v2) after authenticating via AppRole
func (ss secretsClient) PutSecret(ctx context.Context, path string, data map[string]interface{}) error {
	authInfo, err := ss.vc.Auth().Login(ctx, ss.auth)
	if err != nil {
		return fmt.Errorf("unable to login to AppRole approle method: %w", err)
	}

	if authInfo == nil {
		return fmt.Errorf("no approle info was returned after login")
	}

	data = map[string]interface{}{"data": data}

	_, werr := ss.vc.Logical().Write(path, data)
	if werr != nil {
		return fmt.Errorf("unable to write secret: %w", werr)
	}

	log.Println("put secret")
	return nil
}

// GetDatabaseCredentials retrieves a new set of temporary database credentials from Vault
func (ss secretsClient) GetDatabaseCredentials(ctx context.Context) (*DatabaseCredentials, error) {
	// TODO: move all of these Logins to NewVaultAppRoleClient and kick off goroutine with LifetimeWatcher
	authInfo, err := ss.vc.Auth().Login(ctx, ss.auth)
	if err != nil {
		return nil, fmt.Errorf("unable to login to AppRole approle method: %w", err)
	}
	if authInfo == nil {
		return nil, fmt.Errorf("no approle info was returned after login")
	}

	path := env.GetOrDefault(env.VaultDBCredsPath, "database/creds/dev-readonly")
	secret, err := ss.vc.Logical().Read(path)
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
