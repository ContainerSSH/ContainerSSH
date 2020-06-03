package main

import (
	"encoding/json"
	"golang.org/x/crypto/ssh"
	"log"
	"net/http"
)

type passwordAuthRequest struct {
	User          string   `json:"user"`
	RemoteAddress string   `json:"remoteAddress"`
	SessionId     string   `json:"sessionIdBase64"`
	Password      string   `json:"passwordBase64"`
}

type publicKeyAuthRequest struct {
	User          string   `json:"user"`
	RemoteAddress string   `json:"remoteAddress"`
	SessionId     string   `json:"sessionIdBase64"`
	// serialized key data in SSH wire format
	PublicKey string `json:"publicKeyBase64"`
}

type authResponse struct {
	Success     bool            `json:"success"`
	Permissions ssh.Permissions `json:"permissions"`
}

func authPassword(w http.ResponseWriter, req *http.Request) {
	var authRequest passwordAuthRequest
	err := json.NewDecoder(req.Body).Decode(&authRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	authResponse := authResponse{
		Success: false,
		Permissions: ssh.Permissions{
			CriticalOptions: map[string]string{},
			Extensions: map[string]string{},
		},
	}
	if authRequest.User == "foo" {
		authResponse.Success = true
	}

	json.NewEncoder(w).Encode(authResponse)
}

func authPublicKey(w http.ResponseWriter, req *http.Request) {
	var authRequest publicKeyAuthRequest
	err := json.NewDecoder(req.Body).Decode(&authRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Print(authRequest)

	authResponse := authResponse{
		Success: false,
		Permissions: ssh.Permissions{
			CriticalOptions: map[string]string{},
			Extensions: map[string]string{},
		},
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
