package main

import (
	"testing"

	"github.com/hashicorp/go-hclog"
	kv "github.com/hashicorp/vault-plugin-secrets-kv"
	"github.com/hashicorp/vault/api"
	vaulthttp "github.com/hashicorp/vault/http"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/hashicorp/vault/vault"
)

var (
	testCluster *vault.TestCluster
	testClient  *api.Client
)

func createVaultTestCluster(t *testing.T) *vault.TestCluster {
	t.Helper()
	coreConfig := &vault.CoreConfig{
		Logger: hclog.NewNullLogger(),
		LogicalBackends: map[string]logical.Factory{
			"kv": kv.Factory,
		},
	}
	cluster := vault.NewTestCluster(t, coreConfig, &vault.TestClusterOptions{
		HandlerFunc: vaulthttp.Handler,
	})
	cluster.Start()

	// Create KV V2 mount
	if err := cluster.Cores[0].Client.Sys().Mount("kv-v2", &api.MountInput{
		Type: "kv",
		Options: map[string]string{
			"version": "2",
		},
	}); err != nil {
		t.Fatal(err)
	}
	return cluster
}

func beforeEach(t *testing.T) {
	t.Helper()
	testCluster = createVaultTestCluster(t)
	testCluster.UnsealCores(t)
	testClient = testCluster.Cores[0].Client
}

func TestGetSecretAPIKey(t *testing.T) {
	beforeEach(t)
	defer testCluster.Cleanup()

	var expectedAPIKey = "testAPIKey"

	v := &Vault{
		client: testClient,
		parameters: VaultParameters{
			apiKeyPath:  "kv-v2/data/api-key",
			apiKeyField: "api-key-field",
		},
	}

	secret := map[string]interface{}{
		"data":     map[string]interface{}{v.parameters.apiKeyField: expectedAPIKey},
		"metadata": map[string]interface{}{"version": 2},
	}

	// seed test vault with expected key
	_, err := testClient.Logical().Write(v.parameters.apiKeyPath, secret)
	if err != nil {
		t.Fatalf("unable to seed test cluster %s", err)
	}

	actualAPIKey, err := v.GetSecretAPIKey()

	if err != nil {
		t.Fatalf("unable to get secret %s", err)
	}

	if actualAPIKey != expectedAPIKey {
		t.Fail()
	}

}
