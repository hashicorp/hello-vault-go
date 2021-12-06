package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/hashicorp/hello-vault-go/clients"
	"github.com/hashicorp/hello-vault-go/env"
	"github.com/hashicorp/hello-vault-go/models"
)

const (
	apiKeyPath = "kv-v2/data/api-key"
)

var (
	client = http.Client{
		Timeout: time.Second * 10,
	}
)

type AppHandler struct {
	DB      *sql.DB
	Secrets *clients.SecretsClient
}

func ListenAndServe(h AppHandler) {
	r := mux.NewRouter()
	r.StrictSlash(true)
	setRoutes(r, h)

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

// setRoutes adds handler functions to the router for specific route / method pairs
func setRoutes(r *mux.Router, h AppHandler) {
	// Product handlers using configured database connection
	r.HandleFunc("/products", h.getProducts()).Methods("GET")

	// Retrieve api key from vault to create an authenticated request (read from vault)
	r.HandleFunc("/payments", h.createPayment()).Methods("POST")

	// Update api key used for making payments (write to vault)
	r.HandleFunc("/admin/keys", h.updateAPIKey()).Methods("PUT")
}

// APIUpdateRequest is the shape of the request for updating the API key
type APIUpdateRequest struct {
	Key string `json:"key"`
}

func (h AppHandler) getProducts() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		products, err := models.GetAllProducts(h.DB)
		if err != nil {
			ErrorResponder(err, w, r)
			return
		}
		JSONResponder(http.StatusOK, products, w, r)
	}
}

func (h AppHandler) createPayment() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// retrieve secret from Vault passing in the active context and path to secret
		secret, err := h.Secrets.GetSecret(r.Context(), apiKeyPath)
		if err != nil {
			ErrorResponder(err, w, r)
			return
		}

		// check that our expected key is in the returned secret
		apiKey, ok := secret["apiKey"]
		if !ok {
			ErrorResponder(fmt.Errorf("key apiKey not in secret"), w, r)
			return
		}

		req, err := http.NewRequest("GET", env.GetOrDefault(env.SecureServer, "http://localhost:1717/api"), nil)
		if err != nil {
			ErrorResponder(err, w, r)
			return
		}

		req.Header.Set("X-API-KEY", apiKey.(string))

		resp, err := client.Do(req)
		if err != nil {
			ErrorResponder(err, w, r)
			return
		}
		defer resp.Body.Close()

		body := make(map[string]interface{})
		err = json.NewDecoder(resp.Body).Decode(&body)
		if err != nil {
			ErrorResponder(err, w, r)
			return
		}

		JSONResponder(http.StatusOK, body, w, r)
	}
}

func (h AppHandler) updateAPIKey() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		p := &APIUpdateRequest{}
		err := json.NewDecoder(r.Body).Decode(p)
		if err != nil {
			ErrorResponder(err, w, r)
			return
		}

		data := make(map[string]interface{})
		data["apiKey"] = p.Key

		err = h.Secrets.PutSecret(r.Context(), apiKeyPath, data)
		if err != nil {
			ErrorResponder(err, w, r)
			return
		}
		JSONResponder(http.StatusNoContent, nil, w, r)
	}
}
