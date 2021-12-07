package main

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

type DatabaseParamters struct {
	hostname string
	port     string
	name     string
	timeout  time.Duration
}

type Database struct {
	connection      *sql.DB
	connectionMutex sync.Mutex
	parameters      DatabaseParamters
	vault           *Vault
}

func NewDatabase(ctx context.Context, parameters DatabaseParamters, vault *Vault) (*Database, error) {
	database := Database{
		parameters:      parameters,
		connection:      nil,
		connectionMutex: sync.Mutex{},
		vault:           vault,
	}

	// establish the first connection
	if err := database.Reconnect(ctx); err != nil {
		return nil, err
	}

	return &database, nil
}

func (db *Database) Close() error {
	return db.Close()
}

// Reconnect will be called periodically to refresh the database connection
// since the dynamic credentials expire after some time, it will:
//   1. construct a connection string using dynamic credentials from Vault
//   2. establishe a database connection
//   3. overwrite the existing connection with the new one behind a mutex
func (db *Database) Reconnect(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, db.parameters.timeout)
	defer cancel()

	credentials, err := db.vault.GetDatabaseCredentials(ctx)
	if err != nil {
		return fmt.Errorf("unable to get database credentials: %w", err)
	}

	connectionString := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		db.parameters.hostname,
		db.parameters.port,
		db.parameters.name,
		credentials.Username,
		credentials.Password,
	)

	connection, err := sql.Open("postgres", connectionString)
	if err != nil {
		return fmt.Errorf("unable to open database connection: %w", err)
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
			return fmt.Errorf("failed to successfully ping database before context timeout: %w", err)
		}
	}

	// protect the connection swap with a mutex to avoid potential race conditions
	db.connectionMutex.Lock()
	db.connection = connection
	db.connectionMutex.Unlock()

	return nil
}

type Product struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// GetProducts is a simple query function to demonstrate that we have
// successfully established a database connection with the Vault credentials.
func (db *Database) GetProducts(ctx context.Context) ([]Product, error) {
	/* */ db.connectionMutex.Lock()
	defer db.connectionMutex.Unlock()

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
