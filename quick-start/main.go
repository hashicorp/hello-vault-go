package main

import (
	"fmt"
	"log"

	vault "github.com/hashicorp/vault/api"
)

const (
	vaultServerAddr = "http://127.0.0.1:8200"
	token           = "dev-only-token" // This is the root token for our Vault dev server. Don't do this in production!
	pathToSecret    = "secret/data/my-secret-password"
	secretKey       = "password"
	secretValue     = "Hashi123"
)

func main() {
	config := vault.DefaultConfig()

	config.Address = vaultServerAddr

	client, err := vault.NewClient(config)
	if err != nil {
		log.Fatalf("unable to initialize Vault client: %v", err)
	}

	// Authentication
	client.SetToken(token)

	secretData := map[string]interface{}{
		"data": map[string]interface{}{
			secretKey: secretValue,
		},
	}

	// Writing a secret
	_, err = client.Logical().Write(pathToSecret, secretData)
	if err != nil {
		log.Fatalf("unable to write secret: %v", err)
	}

	fmt.Println("Secret written successfully.")

	// Reading a secret
	secret, err := client.Logical().Read(pathToSecret)
	if err != nil {
		log.Fatalf("unable to read secret: %v", err)
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		log.Fatalf("data type assertion failed: %T %#v", secret.Data["data"], secret.Data["data"])
	}

	value, ok := data[secretKey].(string)
	if !ok {
		log.Fatalf("value type assertion failed: %T %#v", data[secretKey], data[secretKey])
	}

	if value != secretValue {
		log.Fatalf("unexpected password value %q retrieved from vault", value)
	}

	fmt.Println("Access granted!")
}
