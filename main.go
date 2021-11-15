package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/hashicorp/hello-vault-go/handlers"
	"github.com/hashicorp/hello-vault-go/models"
	_ "github.com/lib/pq"
)

func main() {
	initDatabaseConnection()
	initRouter()
}

// sets up a database connection
func initDatabaseConnection() {
	// TODO: convert this to use dynamic DB credentials from Vault
	hostPort := 5432
	hostName := "db"
	user := "tmptmp"
	password := "temp"
	dbName := "postgres"
	connectionStr := fmt.Sprintf("port=%d host=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		hostPort, hostName, user, password, dbName)

	var err error
	models.DB, err = sql.Open("postgres", connectionStr)
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}

	tryDatabase(models.DB)
}

// sets up router
func initRouter() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", handlers.Hello).Methods("GET")
	router.HandleFunc("/products", handlers.GetProducts).Methods("GET")
	http.Handle("/", router)

	// start listening to requests
	fmt.Println("Listening on port 8080.")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func tryDatabase(db *sql.DB) {
	var err error
	for i := 0; i < 30; i++ {
		err = db.Ping()
		if err == nil {
			fmt.Println("Connected to database successfully.")
			return
		}
		log.Print("Database ping failed. Retrying in 1s")
		time.Sleep(time.Second)
	}
}
