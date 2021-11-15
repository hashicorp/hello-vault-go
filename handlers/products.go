package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/hello-vault-go/models"
)

type Response struct {
	Products []models.Product `json:"products"`
}

func GetProducts(w http.ResponseWriter, r *http.Request) {
	var response Response

	products, err := models.GetAllProducts()
	if err != nil {
		// TODO: clean error response
		w.Write([]byte(fmt.Sprintf("{\"error\": \"%s\"}", err.Error())))
		return
	}

	response.Products = products

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		// TODO: clean error response
		w.Write([]byte(fmt.Sprintf("{\"error\": \"%s\"}", err.Error())))
		return
	}

	w.Write(jsonResponse)
}
