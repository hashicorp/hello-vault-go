package clients

import (
	"database/sql"
	"fmt"
	"github.com/hashicorp/hello-vault-go/util"
	_ "github.com/lib/pq"
	"log"
)

const (
	EnvDBHost = "DB_HOST"
	EnvDBPort = "DB_PORT"
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
	hostName := util.GetEnvOrDefault(EnvDBHost, "localhost")
	hostPort := util.GetEnvOrDefault(EnvDBPort, "5432")
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
