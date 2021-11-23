package env

import (
	"log"
	"os"
)

const (
	ServerAddress = "SERVER_ADDRESS"
	ServerPort    = "SERVER_PORT"

	DBHost = "DB_HOST"
	DBPort = "DB_PORT"

	SecretID     = "SECRET_ID"
	AppRoleID    = "APPROLE_ROLE_ID"
	VaultAddress = "VAULT_ADDRESS"
)

// GetEnvOrDefault retrieves the value of the environment variable named by the key, if unset returns default
func GetEnvOrDefault(key, def string) string {
	v, exists := os.LookupEnv(key)
	if !exists {
		return def
	}
	return v
}

// MustGetEnv retrieves the value of the environment variable named by the key, if unset calls log.Fatal()
func MustGetEnv(key string) string {
	v, exists := os.LookupEnv(key)
	if !exists {
		log.Fatalf("%s not found in environment", key)
	}
	return v
}
