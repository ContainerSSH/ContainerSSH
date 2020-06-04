package main

import (
	"containerssh/auth"
	"encoding/json"
	"log"
	"net/http"
)

func authPassword(w http.ResponseWriter, req *http.Request) {
	var authRequest auth.PasswordRequest
	err := json.NewDecoder(req.Body).Decode(&authRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	authResponse := auth.Response{
		Success: false,
	}
	if authRequest.User == "foo" {
		authResponse.Success = true
	}

	json.NewEncoder(w).Encode(authResponse)
}

func authPublicKey(w http.ResponseWriter, req *http.Request) {
	var authRequest auth.PublicKeyRequest
	err := json.NewDecoder(req.Body).Decode(&authRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Print(authRequest)

	authResponse := auth.Response{
		Success: false,
	}
	if authRequest.User == "foo" {
		authResponse.Success = true
	}

	json.NewEncoder(w).Encode(authResponse)
}

func main() {
	http.HandleFunc("/pubkey", authPublicKey)
	http.HandleFunc("/password", authPassword)
	http.ListenAndServe(":8080", nil)
}
