package models

import (
	"database/sql"

	"github.com/hashicorp/hello-vault-go/clients"
	"github.com/hashicorp/hello-vault-go/util"
)

var db = clients.MustGetDatabase()

type Product struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func GetAllProducts() ([]Product, error) {
	rows, err := db.Query("SELECT * FROM products;")
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, util.NotFoundError
		}
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
