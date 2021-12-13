package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	database             *Database
	vault                *Vault
	secureServiceAddress string
}

// (POST /payments) : demonstrates fetching a static secret from Vault and using it to talk to another service
func (h *Handlers) Payments(c *gin.Context) {
	// retrieve the secret from Vault
	secret, err := h.vault.GetSecretAPIKey(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// check that our expected key is in the returned secret
	apiKey, ok := secret["apiKey"]
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "the secret retrieved from vault is missing 'apiKey' field"})
		return
	}

	request, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, h.secureServiceAddress, nil)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// use the api key in our request header
	request.Header.Set("X-API-KEY", apiKey.(string))

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer func() {
		_ = response.Body.Close()
	}()

	// forward the response back to the caller
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("could not read secure service response body: %v", err)})
		return
	}

	c.Data(response.StatusCode, "application/json", b)
}

// (GET /products) : demonstrates database authentication with dynamic secrets
func (h *Handlers) Products(c *gin.Context) {
	products, err := h.database.GetProducts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, products)
}
