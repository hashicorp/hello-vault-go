package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type Database struct {
	connection *sql.DB
	vault      *Vault
}

// NewDatabase constructs a connection string using dynamic credentials from vault client & establishes a database connection
func NewDatabase(ctx context.Context, hostname, port, name string, timeout time.Duration, vault *Vault) (*Database, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	credentials, err := vault.GetDatabaseCredentials(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get database credentials: %w", err)
	}

	connectionStr := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		hostname,
		port,
		name,
		credentials.Username,
		credentials.Password,
	)

	connection, err := sql.Open("postgres", connectionStr)
	if err != nil {
		return nil, fmt.Errorf("unable to open database connection: %w", err)
	}

	// wait until the database is ready or timeout expires
	for {
		err = connection.Ping()
		if err == nil {
			break
		}
		select {
		case <-time.After(500 * time.Millisecond):
			continue
		case <-ctx.Done():
			return nil, fmt.Errorf("failed to successfully ping database before context timeout: %w", err)
		}
	}

	return &Database{
		connection: connection,
		vault:      vault,
	}, nil
}

func (db *Database) Close() error {
	return db.Close()
}

type Product struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (db *Database) GetProducts(ctx context.Context) ([]Product, error) {
	const query = "SELECT id, name FROM products"

	rows, err := db.connection.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute %q query: %w", query, err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(
			&p.ID,
			&p.Name,
		); err != nil {
			return nil, fmt.Errorf("failed to scan table row for %q query: %w", query, err)
		}
		products = append(products, p)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error after scanning %q query: %w", query, err)
	}

	return products, nil
}
