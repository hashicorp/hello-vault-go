package main

//
// The tests below assume that the docker-compose environment is up. Either
// bring it up manually with docker compose up -d or use the helper script:
//
// ./run-tests.sh
//

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const appAddress = "http://localhost:8080"

func TestPostPayments(t *testing.T) {
	sendAndVerify(
		t,
		http.MethodPost,
		fmt.Sprintf("%s/payments", appAddress),
		`{"message":"hello world!"}`,
	)
}

func TestGetProducts(t *testing.T) {
	sendAndVerify(
		t,
		http.MethodGet,
		fmt.Sprintf("%s/products", appAddress),
		`[{"id":1,"name":"Rustic Webcam"},{"id":2,"name":"Haunted Coloring Book"}]`,
	)
}

func sendAndVerify(t *testing.T, method, endpoint, expected string) {
	t.Logf("%s %s: sending request", method, endpoint)

	request, err := http.NewRequest(method, endpoint, nil)
	require.NoErrorf(t, err, "%s %s: could not form the request", method, endpoint)

	response, err := http.DefaultClient.Do(request)
	require.NoErrorf(t, err, "%s %s: could not send the request", method, endpoint)
	defer func() {
		_ = response.Body.Close()
	}()

	assert.Equalf(t, http.StatusOK, response.StatusCode, "%s %s: the request returned non-200 response", method, endpoint)

	body, err := ioutil.ReadAll(response.Body)
	require.NoError(t, err, "%s %s: couldn't read the response body", method, endpoint)

	assert.Equalf(t, expected, string(body), "%s %s: the response body did not match the expected one", method, endpoint)
}
