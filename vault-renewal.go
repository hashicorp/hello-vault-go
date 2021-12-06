package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	vault "github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/approle"
)

// Once you've set the token for your Vault client, you will need to
// periodically renew its lease.
//
// A function like this should be run as a goroutine to avoid blocking.
//
// Production applications may also wish to be more tolerant of failures and
// retry rather than exiting.
//
// Additionally, enterprise Vault users should be aware that due to eventual
// consistency, the API may return unexpected errors when running Vault with
// performance standbys or performance replication, despite the client having
// a freshly renewed token. See https://www.vaultproject.io/docs/enterprise/consistency#vault-1-7-mitigations
// for several ways to mitigate this which are outside the scope of this code sample.
func (v *Vault) RenewVaultLogin() {
	for {
		vaultLoginResp, err := login(v.client)
		if err != nil {
			log.Fatalf("unable to authenticate to Vault: %v", err)
		}
		tokenErr := manageTokenLifecycle(v.client, vaultLoginResp)
		if tokenErr != nil {
			log.Fatalf("unable to start managing token lifecycle: %v", tokenErr)
		}
	}
}

func (v *Vault) RenewDatabaseLogin(db *sql.DB) {
	for {
		// TODO: reconnect to DB altogether
		var dbCreds *vault.Secret

		credsErr := manageDBSecretLifecycle(v.client, dbCreds)
		if credsErr != nil {
			log.Fatalf("unable to start managing database credentials lifecycle: %v", credsErr)
		}
	}
}

func login(client *vault.Client) (*vault.Secret, error) {
	// A combination of a Role ID and Secret ID is required to log in to Vault
	// with an AppRole. We're passing this in from an environment variable, "APPROLE_ROLE_ID".
	role := os.Getenv("VAULT_APPROLE_ROLE_ID")

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

	authInfo, err := client.Auth().Login(context.TODO(), appRoleAuth)
	if err != nil {
		return nil, fmt.Errorf("unable to login to AppRole approle method: %w", err)
	}
	if authInfo == nil {
		return nil, fmt.Errorf("no approle info was returned after login")
	}

	return authInfo, nil
}

// Starts token lifecycle management. Returns only fatal errors as errors,
// otherwise returns nil so we can attempt login again.
func manageTokenLifecycle(client *vault.Client, token *vault.Secret) error {
	renew := token.Auth.Renewable // You may notice a different top-level field called Renewable. That one is used for dynamic secrets renewal, not token renewal.
	if !renew {
		log.Printf("Token is not configured to be renewable. Re-attempting login.")
		return nil
	}

	watcher, err := client.NewLifetimeWatcher(&vault.LifetimeWatcherInput{
		Secret:    token,
		Increment: 3600, // Learn more about this optional value in https://www.vaultproject.io/docs/concepts/lease#lease-durations-and-renewal
	})
	if err != nil {
		return fmt.Errorf("unable to initialize new lifetime watcher for renewing auth token: %w", err)
	}

	go watcher.Start()
	defer watcher.Stop()

	for {
		select {
		// `DoneCh` will return if renewal fails, or if the remaining lease
		// duration is under a built-in threshold and either renewing is not
		// extending it or renewing is disabled. In any case, the caller
		// needs to attempt to log in again.
		case err := <-watcher.DoneCh():
			if err != nil {
				log.Printf("Failed to renew token: %v. Re-attempting login.", err)
				return nil
			}
			log.Printf("Token can no longer be renewed. Re-attempting login.")
			return nil

		// Successfully completed renewal
		case renewal := <-watcher.RenewCh():
			log.Printf("Successfully renewed auth token: %#v", renewal)
		}
	}
}

// Starts DB credential lifecycle management. Returns only fatal errors as errors,
// otherwise returns nil so we can attempt login again.
func manageDBSecretLifecycle(client *vault.Client, creds *vault.Secret) error {
	// TODO: Make sure that this is the correct field for checking DB secret renewableness.
	// It's actually the nested Auth.Renewable field for auth token secrets, but I was told that dynamic secrets use a different field.
	renew := creds.Renewable
	if !renew {
		log.Printf("Database creds secret is not configured to be renewable. Re-attempting database connection.")
		return nil
	}

	watcher, err := client.NewLifetimeWatcher(&vault.LifetimeWatcherInput{
		Secret:    creds,
		Increment: 3600, // Learn more about this optional value in https://www.vaultproject.io/docs/concepts/lease#lease-durations-and-renewal
	})
	if err != nil {
		return fmt.Errorf("unable to initialize new lifetime watcher for renewing database credentials: %w", err)
	}

	go watcher.Start()
	defer watcher.Stop()

	for {
		select {
		// `DoneCh` will return if renewal fails, or if the remaining lease
		// duration is under a built-in threshold and either renewing is not
		// extending it or renewing is disabled. In any case, the caller
		// needs to fetch a brand-new set of creds again rather than renewing.
		case err := <-watcher.DoneCh():
			if err != nil {
				log.Printf("Failed to renew current database credentials: %v. Re-attempting database connection.", err)
				return nil
			}
			log.Printf("Database credentials can no longer be renewed. Re-attempting database connection.")
			return nil

		// Successfully completed renewal
		case renewal := <-watcher.RenewCh():
			log.Printf("Successfully renewed database credentials: %#v", renewal)
		}
	}
}
