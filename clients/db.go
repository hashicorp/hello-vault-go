package clients

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"

	"github.com/hashicorp/hello-vault-go/env"
)

func MustGetDatabase() *sql.DB {
	db, err := GetDatabase()
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func GetDatabase() (*sql.DB, error) {
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

	err = db.Ping()

	return db, err
}
