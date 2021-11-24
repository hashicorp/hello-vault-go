package env

import (
	"log"
	"os"
)

const (
	// ServerAddress is the key for the environment variable that holds the listener address for the API
	ServerAddress = "SERVER_ADDRESS"
	// ServerPort is the key for the environment variable that holds the listener port for the API
	ServerPort    = "SERVER_PORT"

	// DBHost is the key for the environment variable that holds the db host address
	DBHost = "DB_HOST"
	// DBPort is the key for the environment variable that holds the db port
	DBPort = "DB_PORT"

	// SecretID is the key for the environment variable that holds the secret ID used to authenticate Vault requests
	SecretID     = "SECRET_ID"
	// AppRoleID is the key for the environment variable that holds the AppRoleID used to authenticate Vault requests
	AppRoleID    = "APPROLE_ROLE_ID"
	// VaultAddress is the key for the environment variable that holds the address to Vault
	VaultAddress = "VAULT_ADDRESS"

	// SecureServer is the key for the environment variable that holds the address to our simulated secure server
	SecureServer = "SECURE_ADDRESS"
)

// GetOrDefault retrieves the value of the environment variable named by the key, if unset returns default
func GetOrDefault(key, def string) string {
	v, exists := os.LookupEnv(key)
	if !exists {
		return def
	}
	return v
}

// MustGet retrieves the value of the environment variable named by the key, if unset calls log.Fatal()
func MustGet(key string) string {
	v, exists := os.LookupEnv(key)
	if !exists {
		log.Fatalf("%s not found in environment", key)
	}
	return v
}
