package util

import (
	"log"
	"os"
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
