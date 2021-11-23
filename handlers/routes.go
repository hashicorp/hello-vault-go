package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/hashicorp/hello-vault-go/clients"
	"github.com/hashicorp/hello-vault-go/models"
	"github.com/hashicorp/hello-vault-go/util"
)

const (
	apiKeyPath = "kv-v2/data/api-key"
)

var (
	// database connection with authentication managed by Vault
	db = clients.MustGetDatabase()
	// sample secret store backed by Vault
	ss = clients.MustMakeNewSecretStore()

	client = http.Client{
		Timeout:       time.Second*10,
	}
)

func SetRoutes(r *mux.Router) {
	// Product handlers using configured database connection
	r.HandleFunc("/products", getProducts()).Methods("GET")

	// Retrieve api key from vault to create an authenticated request (read from vault)
	r.HandleFunc("/payment", createPayment()).Methods("POST")

	// Update api key used for making payments (write to vault)
	r.HandleFunc("/admin/keys", updateAPIKey()).Methods("PUT")
}

type APIUpdateRequest struct {
	Key string `json:"key"`
}

func updateAPIKey() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		p := &APIUpdateRequest{}
		err := json.NewDecoder(r.Body).Decode(p)
		if err != nil {
			util.ErrorResponder(err, w, r)
			return
		}

		var data map[string]interface{}
		data["apiKey"] = p.Key

		err = ss.PutSecret(r.Context(), apiKeyPath, data)
		if err != nil {
			util.ErrorResponder(err, w, r)
			return
		}
		util.JSONResponder(http.StatusNoContent, nil, w, r)
	}
}

func createPayment() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		//retrieve secret from Vault passing in the active context and path to secret
		secret, err := ss.GetSecret(r.Context(), apiKeyPath)
		if err != nil {
			util.ErrorResponder(err, w, r)
			return
		}

		//check that our expected key is in the returned secret
		apiKey, ok := secret["apiKey"]
		if !ok {
			util.ErrorResponder(fmt.Errorf("key apiKey not in secret"), w, r)
			return
		}

		req, err := http.NewRequest("GET", "https://postman-echo.com/headers", nil)
		if err != nil {
			util.ErrorResponder(err, w, r)
			return
		}

		req.Header.Set("API_KEY", apiKey.(string))

		resp, err := client.Do(req)
		if err != nil {
			util.ErrorResponder(err, w, r)
			return
		}
		defer resp.Body.Close()

		body := make(map[string]interface{})
		err = json.NewDecoder(resp.Body).Decode(&body)
		if err != nil {
			util.ErrorResponder(err, w, r)
			return
		}

		util.JSONResponder(http.StatusOK, body, w, r)
	}
}

func getProducts() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		products, err := models.GetAllProducts()
		if err != nil {
			util.ErrorResponder(err, w, r)
			return
		}
		util.JSONResponder(http.StatusOK, products, w, r)
	}
}