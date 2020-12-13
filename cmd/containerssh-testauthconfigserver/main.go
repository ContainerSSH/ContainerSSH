// ContainerSSH Authentication and Configuration Server
//
// This OpenAPI document describes the API endpoints that are required for implementing an authentication
// and configuration server for ContainerSSH. (See https://github.com/containerssh/containerssh for details.)
//
//     Schemes: http, https
//     Host: localhost
//     BasePath: /
//     Version: 0.3.0
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
	"log"
	"net/http"
	"os"

	"github.com/containerssh/configuration"
	"github.com/containerssh/structutils"

	"github.com/containerssh/containerssh/test/protocol"
)

type authConfigServer struct {
}

func (s *authConfigServer) authPassword(w http.ResponseWriter, req *http.Request) {
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

	authResponse := protocol.AuthResponse{
		Success: false,
	}
	if authRequest.Username == "foo" || authRequest.Username == "busybox" {
		authResponse.Success = true
	}

	_ = json.NewEncoder(w).Encode(authResponse)
}

func (s *authConfigServer) authPublicKey(w http.ResponseWriter, req *http.Request) {
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

	authResponse := protocol.AuthResponse{
		Success: false,
	}
	if authRequest.Username == "foo" || authRequest.Username == "busybox" {
		authResponse.Success = true
	}

	_ = json.NewEncoder(w).Encode(authResponse)
}

func (s *authConfigServer) configHandler(w http.ResponseWriter, req *http.Request) {
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

	defaultConfig := configuration.AppConfig{}
	structutils.Defaults(&defaultConfig)

	response := protocol.ConfigResponse{
		Config: defaultConfig,
	}

	if configRequest.Username == "busybox" {
		response.Config.DockerRun.Config.ContainerConfig.Image = "busybox"
	}

	_ = json.NewEncoder(w).Encode(response)
}

func main() {
	s := &authConfigServer{}
	http.HandleFunc("/pubkey", s.authPublicKey)
	http.HandleFunc("/password", s.authPassword)
	http.HandleFunc("/config", s.configHandler)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
