package util

import (
	"log"
	"os"
)

// GetDefault retrieves the value of the environment variable named by the key, if unset returns default
func GetDefault(key, def string) string {
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
