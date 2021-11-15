package handlers

import "net/http"

func Hello(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello Vault! :)\n"))
}
