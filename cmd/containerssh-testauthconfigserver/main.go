// ContainerSSH Authentication and Configuration Server
//
// This OpenAPI document describes the API endpoints that are required for implementing an authentication
// and configuration server for ContainerSSH. (See https://github.com/janoszen/containerssh for details.)
//
//     Schemes: http, https
//     Host: localhost
//     BasePath: /
//     Version: 0.2.1
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
// swagger:meta
package main

import (
	"encoding/json"
	"github.com/janoszen/containerssh/config"
	"github.com/janoszen/containerssh/protocol"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func authPassword(w http.ResponseWriter, req *http.Request) {
	// swagger:operation POST /password Authentication authPassword
	//
	// Password authentication
	//
	// ---
	// parameters:
	// - name: request
	//   in: body
	//   description: The authentication request
	//   required: true
	//   schema:
	//     "$ref": "#/definitions/PasswordAuthRequest"
	// responses:
	//   "200":
	//     "$ref": "#/responses/AuthResponse"

	var authRequest protocol.PasswordAuthRequest
	err := json.NewDecoder(req.Body).Decode(&authRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Tracef("Password authentication request for user %s", authRequest.User)

	authResponse := protocol.AuthResponse{
		Success: false,
	}
	if authRequest.User == "foo" || authRequest.User == "busybox" {
		authResponse.Success = true
	}

	_ = json.NewEncoder(w).Encode(authResponse)
}

func authPublicKey(w http.ResponseWriter, req *http.Request) {
	// swagger:operation POST /pubkey Authentication authPubKey
	//
	// Public key authentication
	//
	// ---
	// parameters:
	// - name: request
	//   in: body
	//   description: The authentication request
	//   required: true
	//   schema:
	//     "$ref": "#/definitions/PublicKeyAuthRequest"
	// responses:
	//   "200":
	//     "$ref": "#/responses/AuthResponse"

	var authRequest protocol.PublicKeyAuthRequest
	err := json.NewDecoder(req.Body).Decode(&authRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Tracef("Public key authentication request for user %s", authRequest.Username)

	authResponse := protocol.AuthResponse{
		Success: false,
	}
	if authRequest.Username == "foo" || authRequest.Username == "busybox" {
		authResponse.Success = true
	}

	_ = json.NewEncoder(w).Encode(authResponse)
}

func configHandler(w http.ResponseWriter, req *http.Request) {
	// swagger:operation POST /config Configuration getUserConfiguration
	//
	// Fetches the configuration for a user/session
	//
	// ---
	// parameters:
	// - name: request
	//   in: body
	//   description: The configuration request
	//   schema:
	//     "$ref": "#/definitions/ConfigRequest"
	// responses:
	//   "200":
	//     "$ref": "#/responses/ConfigResponse"

	var configRequest protocol.ConfigRequest
	err := json.NewDecoder(req.Body).Decode(&configRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	defaultConfig := &config.AppConfig{}

	response := protocol.ConfigResponse{
		Config: *defaultConfig,
	}

	log.Tracef("Config request for user %s", configRequest.Username)

	if configRequest.Username == "busybox" {
		response.Config.DockerRun.Config.ContainerConfig.Image = "busybox"
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Println(err)
	}
}

func main() {
	log.SetLevel(log.TraceLevel)
	http.HandleFunc("/pubkey", authPublicKey)
	http.HandleFunc("/password", authPassword)
	http.HandleFunc("/config", configHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
