package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/hashicorp/hello-vault-go/internal/env"
	"github.com/hashicorp/hello-vault-go/internal/handlers"
)

func main() {
	r := mux.NewRouter()
	r.StrictSlash(true)
	handlers.SetRoutes(r)

	addr := fmt.Sprintf("%s:%s",
		env.GetOrDefault(env.ServerAddress, "0.0.0.0"),
		env.GetOrDefault(env.ServerPort, "8080"))

	log.Println("starting server at", addr)

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("shutting down the server: %s", err)
	}

}
