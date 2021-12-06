package main

import (
	"time"

	"github.com/hashicorp/hello-vault-go/clients"
	"github.com/hashicorp/hello-vault-go/handlers"
)

func main() {
	// WARNING: The goroutines in this function call log.Fatal if an unrecoverable error is encountered.
	// Production applications may want to add more complex error handling and memory leak protection.

	// create a client that is authenticated with Vault for all future secrets operations
	sc := clients.MustGetSecretsClient()
	// keep Vault connection alive
	go sc.RenewVaultLogin()

	// TODO: Fix the fact that the goroutine is coming back too quickly and so MustGetDatabase fails because the Vault auth hasn't happened yet.
	// Maybe make it so MustGetSecretsClient does one login by itself first? Or something? How do we make this clean and readable?!

	// create a client for connecting with the DB using dynamic credentials from Vault
	db := clients.MustGetDatabase(time.Second*10, sc)
	defer func() {
		_ = db.Close()
	}()
	// keep database connection alive
	go sc.RenewDatabaseLogin(db)

	// init server
	h := handlers.AppHandler{
		DB:      db,
		Secrets: sc,
	}
	handlers.ListenAndServe(h)
}
