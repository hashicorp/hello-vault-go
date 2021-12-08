package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	vault "github.com/hashicorp/vault/api"
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
func (v *Vault) RenewLoginPeriodically(ctx context.Context, authToken *vault.Secret) {
	/* */ log.Println("RenewLoginPeriodically: begin")
	defer log.Println("RenewLoginPeriodically: end")

	currentAuthToken := authToken

	for {
		if err := v.renewTokenUntilMaxTTL(ctx, currentAuthToken, "auth token"); err != nil {
			// break out when shutdown is requested
			if errors.Is(err, context.Canceled) {
				return
			}

			log.Fatalf("Renew error: %v\n", err) // simplified error handling
		}

		// the auth token's lease has expired and needs to be renewed
		t, err := v.login(ctx)
		if err != nil {
			log.Fatalf("Authentication error: %v\n", err) // simplified error handling
		}

		currentAuthToken = t
	}
}

// RenewDatabaseCredentialsPeriodically uses a similar mechnanism to the one
// above in order to keep the database connection alive after the database
// token expires and needs to be renewed.
func (v *Vault) RenewDatabaseCredentialsPeriodically(
	ctx context.Context,
	databaseToken *vault.Secret,
	reconnect func(ctx context.Context, credentials DatabaseCredentials) error,
) {
	/* */ log.Println("RenewDatabaseCredentialsPeriodically: begin")
	defer log.Println("RenewDatabaseCredentialsPeriodically: end")

	for {
		currentDatabaseToken := databaseToken

		for {
			if err := v.renewTokenUntilMaxTTL(ctx, currentDatabaseToken, "database credentials"); err != nil {
				// break out when shutdown is requested
				if errors.Is(err, context.Canceled) {
					return
				}

				log.Fatalf("Renew error: %v\n", err) // simplified error handling
			}

			// database credentials have expired and need to be renewed
			credentials, token, err := v.GetDatabaseCredentials(ctx)
			if err != nil {
				log.Fatalf("Database credentials error: %v\n", err) // simplified error handling
			}

			reconnect(ctx, credentials)

			currentDatabaseToken = token
		}
	}
}

// renewTokenUntilMaxTTL is a blocking helper function that uses LifetimeWatcher to
// periodically renew the given token when it is close to its 'token_ttl'
// expiration time until it reaches its token_max_ttl expiration time.
func (v *Vault) renewTokenUntilMaxTTL(ctx context.Context, token *vault.Secret, label string) error {
	/* */ log.Printf("Renew cycle for %s: started\n", label)
	defer log.Printf("Renew cycle for %s: the token can no longer be renewed\n", label)

	watcher, err := v.client.NewLifetimeWatcher(&vault.LifetimeWatcherInput{
		Secret:    token,
		Increment: 3600, // learn more: https://www.vaultproject.io/docs/concepts/lease#lease-durations-and-renewal
	})
	if err != nil {
		return fmt.Errorf("unable to initialize %s lifetime watcher: %w", label, err)
	}

	go watcher.Start()
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return context.Canceled

		// DoneCh will return if renewal fails, or if the remaining lease
		// duration is under a built-in threshold and either renewing is not
		// extending it or renewing is disabled.  In both cases, the caller
		// should attempt a re-read of the secret. Clients should check the
		// return value of the channel to see if renewal was successful.
		case err := <-watcher.DoneCh():
			if err != nil {
				return fmt.Errorf("%s renewal failed: %w", label, err)
			}

			return nil

		// RenewCh is a channel that receives a message when a successful
		// renewal takes place and includes metadata about the renewal.
		case <-watcher.RenewCh():
			log.Printf("%s: successfully renewed", label)
		}
	}
}
