package main

import (
	"context"
	"fmt"
	"log"

	vault "github.com/hashicorp/vault/api"
)

// Once you've set the token for your Vault client, you will need to
// periodically renew its lease. Likewise, the database credentials lease
// will expire at some point and also needs to be renewed periodically.
//
// A function like this should be run as a goroutine to avoid blocking.
//
// Production applications may also wish to be more tolerant of failures and
// retry rather than exiting.
//
// Additionally, enterprise Vault users should be aware that due to eventual
// consistency, the API may return unexpected errors when running Vault with
// performance standbys or performance replication, despite the client having
// a freshly renewed token. See the link below for several ways to mitigate
// this which are outside the scope of this code sample.
//
// ref: See https://www.vaultproject.io/docs/enterprise/consistency#vault-1-7-mitigations
func (v *Vault) PeriodicallyRenewSecrets(
	ctx context.Context,
	authToken *vault.Secret,
	databaseCredentials *vault.Secret,
	databaseReconnectFunc func(ctx context.Context, credentials DatabaseCredentials) error,
) {
	/* */ log.Println("renew / recreate secrets loop: begin")
	defer log.Println("renew / recreate secrets loop: end")

	currentAuthToken := authToken
	currentDatabaseCredentials := databaseCredentials

	for {
		r, err := v.renewLeases(ctx, currentAuthToken, currentDatabaseCredentials)
		if err != nil {
			log.Fatalf("renew error: %v", err) // simplified error handling
		}

		if r&exitRequested != 0 {
			return
		}

		if r&expiredAuthToken != 0 {
			log.Printf("auth token: can no longer be renewed; will log in again")

			authToken, err := v.login(ctx)
			if err != nil {
				log.Fatalf("login authentication error: %v", err) // simplified error handling
			}

			currentAuthToken = authToken
		}

		if r&expiredDatabaseCredentials != 0 {
			log.Printf("database credentials: can no longer be renewed; will fetch new credentials & reconnect")

			databaseCredentials, databaseCredentialsSecret, err := v.GetDatabaseCredentials()
			if err != nil {
				log.Fatalf("database credentials error: %v", err) // simplified error handling
			}

			databaseReconnectFunc(ctx, databaseCredentials)

			currentDatabaseCredentials = databaseCredentialsSecret
		}
	}
}

type renewResult uint8

const (
	renewError renewResult = 1 << iota
	exitRequested
	expiredAuthToken
	expiredDatabaseCredentials
)

// renewLeases is a blocking helper function that uses LifetimeWatcher
// instances to periodically renew the given secrets when they are close to
// their 'token_ttl' expiration times until one of the secrets is close to its
// 'token_max_ttl' lease expiration time.
func (v *Vault) renewLeases(ctx context.Context, authToken, databaseCredentials *vault.Secret) (renewResult, error) {
	// auth token
	log.Printf("auth token: starting lifetime watcher; lease duration %ds", authToken.LeaseDuration)

	authTokenWatcher, err := v.client.NewLifetimeWatcher(&vault.LifetimeWatcherInput{
		Secret: authToken,
	})
	if err != nil {
		return renewError, fmt.Errorf("unable to initialize auth token lifetime watcher: %w", err)
	}

	go authTokenWatcher.Start()
	defer authTokenWatcher.Stop()

	// database credentials
	log.Printf("database credentials: starting lifetime watcher; lease duration: %ds", databaseCredentials.LeaseDuration)

	databaseCredentialsWatcher, err := v.client.NewLifetimeWatcher(&vault.LifetimeWatcherInput{
		Secret: databaseCredentials,
	})
	if err != nil {
		return renewError, fmt.Errorf("unable to initialize database credentials lifetime watcher: %w", err)
	}

	go databaseCredentialsWatcher.Start()
	defer databaseCredentialsWatcher.Stop()

	//
	for {
		select {
		case <-ctx.Done():
			return exitRequested, nil

		// DoneCh will return if renewal fails, or if the remaining lease
		// duration is under a built-in threshold and either renewing is not
		// extending it or renewing is disabled.  In both cases, the caller
		// should attempt a re-read of the secret. Clients should check the
		// return value of the channel to see if renewal was successful.
		case err := <-authTokenWatcher.DoneCh():
			// Leases created by a token get revoked when the token is revoked.
			return expiredAuthToken | expiredDatabaseCredentials, err

		case err := <-databaseCredentialsWatcher.DoneCh():
			return expiredDatabaseCredentials, err

		// RenewCh is a channel that receives a message when a successful
		// renewal takes place and includes metadata about the renewal.
		case <-authTokenWatcher.RenewCh():
			log.Printf("auth token: successfully renewed")

		case info := <-databaseCredentialsWatcher.RenewCh():
			log.Printf("database credentials: successfully renewed; remaining lease duration: %ds", info.Secret.LeaseDuration)
		}
	}
}
