package clients

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"

	"github.com/hashicorp/hello-vault-go/util"
)

const (
	EnvDBHost = "DB_HOST"
	EnvDBPort = "DB_PORT"
)

func MustGetDatabase() (db *sql.DB) {
	// TODO: convert this to use dynamic DB credentials from Vault
	hostName := util.GetDefault(EnvDBHost, "localhost")
	hostPort := util.GetDefault(EnvDBPort, "5432")
	user := "tmptmp"
	password := "temp"
	dbName := "postgres"
	connectionStr := fmt.Sprintf("port=%s host=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		hostPort, hostName, user, password, dbName)

	db, err := sql.Open("postgres", connectionStr)

	if err != nil {
		log.Fatal("could not create db", err)
	}

	timeout := time.Now().Add(time.Minute*1)
	timedout := true
	for time.Now().Before(timeout) {
		if db.Ping() != nil {
			time.Sleep(time.Second * 3)
			continue
		}
		timedout = false
		break
	}

	if timedout {
		log.Fatal("couldn't reach database")
	}

	return db
}
