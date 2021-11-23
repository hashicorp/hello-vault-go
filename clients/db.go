package clients

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"

	"github.com/hashicorp/hello-vault-go/env"
)

func MustGetDatabase(timeout time.Duration) *sql.DB {
	db, err := GetDatabase(timeout)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func GetDatabase(timeout time.Duration) (*sql.DB, error) {
	// TODO: convert this to use dynamic DB credentials from Vault
	hostName := env.GetOrDefault(env.DBHost, "localhost")
	hostPort := env.GetOrDefault(env.DBPort, "5432")
	user := "tmptmp"
	password := "temp"
	dbName := "postgres"
	connectionStr := fmt.Sprintf("port=%s host=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		hostPort, hostName, user, password, dbName)

	db, err := sql.Open("postgres", connectionStr)

	if err != nil {
		return nil, err
	}

	// wait until DB is ready or timeout expires
	for start := time.Now(); time.Now().Before(start.Add(timeout)); {
		err = db.Ping()
		if err == nil {
			return db, nil
		}
	}

	return db, err
}