package clients

import (
	"context"
	"fmt"
	"log"

	vault "github.com/hashicorp/vault/api"
	auth "github.com/hashicorp/vault/api/auth/approle"

	"github.com/hashicorp/hello-vault-go/env"
)

type kvV2Store struct {
	vc   *vault.Client
	auth *auth.AppRoleAuth
}

// MustGetNewKVv2Store returns a new KVv2 store backed by Vault or calls log.Fatal()
func MustGetNewKVv2Store() (ss *kvV2Store) {
	ss, err := NewKVv2Store()

	if err != nil {
		log.Fatal("could not get secret store", err)
	}
	return
}

// NewKVv2Store returns a new KVv2 store backed by Vault
func NewKVv2Store() (*kvV2Store, error) {
	ss := &kvV2Store{}
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
	secretID := &auth.SecretID{FromEnv: env.SecretID}

	appRoleAuth, err := auth.NewAppRoleAuth(
		role,
		secretID,
		auth.WithWrappingToken(), // Only required if the secret ID is response-wrapped.
	)

	if err != nil {
		return nil, fmt.Errorf("unable to initialize AppRole auth method: %w", err)
	}
	ss.auth = appRoleAuth
	return ss, nil
}

// GetSecret fetches the latest version of a key-value secret (kv-v2) after authenticating via AppRole,
// an auth method used by machines that are unable to use platform-based
// authentication mechanisms like AWS Auth, Kubernetes Auth, etc.
func (ss kvV2Store) GetSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	authInfo, err := ss.vc.Auth().Login(ctx, ss.auth)
	if err != nil {
		return nil, fmt.Errorf("unable to login to AppRole auth method: %w", err)
	}
	if authInfo == nil {
		return nil, fmt.Errorf("no auth info was returned after login")
	}

	secret, err := ss.vc.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read secret: %w", err)
	}

	log.Println("get secret", *secret)

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("malformed secret returned")
	}

	return data, nil
}

// PutSecret creates or overwrites a key-value secret (kv-v2) after authenticating via AppRole
func (ss kvV2Store) PutSecret(ctx context.Context, path string, data map[string]interface{}) error {
	authInfo, err := ss.vc.Auth().Login(ctx, ss.auth)
	if err != nil {
		return fmt.Errorf("unable to login to AppRole auth method: %w", err)
	}

	if authInfo == nil {
		return fmt.Errorf("no auth info was returned after login")
	}

	secret, err := ss.vc.Logical().Write(path, data)
	if err != nil {
		return fmt.Errorf("unable to write secret: %w", err)
	}

	log.Println("put secret", secret)
	return nil
}
