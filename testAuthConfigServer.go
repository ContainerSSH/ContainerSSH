package main

import (
	"containerssh/config/util"
	"containerssh/protocol"
	"encoding/json"
	"log"
	"net/http"
)

func authPassword(w http.ResponseWriter, req *http.Request) {
	var authRequest protocol.PasswordAuthRequest
	err := json.NewDecoder(req.Body).Decode(&authRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	authResponse := protocol.AuthResponse{
		Success: false,
	}
	if authRequest.User == "foo" || authRequest.User == "busybox" {
		authResponse.Success = true
	}

	_ = json.NewEncoder(w).Encode(authResponse)
}

func authPublicKey(w http.ResponseWriter, req *http.Request) {
	var authRequest protocol.PublicKeyAuthRequest
	err := json.NewDecoder(req.Body).Decode(&authRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	authResponse := protocol.AuthResponse{
		Success: false,
	}
	if authRequest.User == "foo" || authRequest.User == "busybox" {
		authResponse.Success = true
	}

	_ = json.NewEncoder(w).Encode(authResponse)
}

func configHandler(w http.ResponseWriter, req *http.Request) {
	var configRequest protocol.ConfigRequest
	err := json.NewDecoder(req.Body).Decode(&configRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	defaultConfig, err := util.GetDefaultConfig()
	if err != nil {
		w.WriteHeader(500)
		return
	}

	response := protocol.ConfigResponse{
		Config: *defaultConfig,
	}

	if configRequest.Username == "busybox" {
		response.Config.DockerRun.Config.ContainerConfig.Image = "busybox"
	}

	_ = json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/pubkey", authPublicKey)
	http.HandleFunc("/password", authPassword)
	http.HandleFunc("/config", configHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
