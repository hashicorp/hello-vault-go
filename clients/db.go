package clients

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"

	"github.com/hashicorp/hello-vault-go/env"
)

func MustGetDatabase(timeout time.Duration, vault *SecretsClient) *sql.DB {
	db, err := GetDatabase(timeout, vault)
	if err != nil {
		log.Fatalf("could not reach database: %s", err)
	}
	return db
}

func GetDatabase(timeout time.Duration, sc *SecretsClient) (*sql.DB, error) {
	hostName := env.GetOrDefault(env.DBHost, "localhost")
	hostPort := env.GetOrDefault(env.DBPort, "5432")
	dbName := env.GetOrDefault(env.DBName, "postgres")

	creds, err := sc.GetDatabaseCredentials(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("unable to get database credentials: %w", err)
	}

	connectionStr := fmt.Sprintf("port=%s host=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		hostPort, hostName, creds.Username, creds.Password, dbName)

	db, err := sql.Open("postgres", connectionStr)

	if err != nil {
		return nil, fmt.Errorf("unable to open database connection: %w", err)
	}

	// wait until DB is ready or timeout expires
	for start := time.Now(); time.Now().Before(start.Add(timeout)); {
		err = db.Ping()
		if err == nil {
			return db, nil
		}
		log.Print("Database ping failed. Retrying.")
	}

	return db, err
}
