package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/hashicorp/hello-vault-go/handlers"
	"github.com/hashicorp/hello-vault-go/util"
)

const (
	EnvServerAddress = "SERVER_ADDRESS"
	EnvServerPort    = "SERVER_PORT"
)

func main() {
	r := mux.NewRouter()
	r.StrictSlash(true)
	handlers.SetRoutes(r)

	addr := fmt.Sprintf("%s:%s",
		util.GetDefault(EnvServerAddress, "0.0.0.0"),
		util.GetDefault(EnvServerPort, "8080"))

	log.Println("starting server at", addr)

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("shutting down the server: %s", err)
	}

}
