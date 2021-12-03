package models

import (
	"time"

	"github.com/hashicorp/hello-vault-go/internal/clients"
)

var db = clients.MustGetDatabase(time.Second * 10)

type Product struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func GetAllProducts() ([]Product, error) {
	rows, err := db.Query("SELECT * FROM products;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		err = rows.Scan(&p.ID, &p.Name)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}
