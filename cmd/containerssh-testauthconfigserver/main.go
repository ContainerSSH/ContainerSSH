// ContainerSSH Authentication and Configuration Server
//
// This OpenAPI document describes the API endpoints that are required for implementing an authentication
// and configuration server for ContainerSSH. (See https://github.com/janoszen/containerssh for details.)
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
	"os"

	"net/http"

	"github.com/janoszen/containerssh/config"
	"github.com/janoszen/containerssh/log"
	"github.com/janoszen/containerssh/log/writer"
	"github.com/janoszen/containerssh/protocol"

)

type authConfigServer struct {
	logger log.Logger
}

func (s * authConfigServer) authPassword(w http.ResponseWriter, req *http.Request) {
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

	s.logger.DebugF("password authentication request for user %s", authRequest.User)

	authResponse := protocol.AuthResponse{
		Success: false,
	}
	if authRequest.User == "foo" || authRequest.User == "busybox" {
		authResponse.Success = true
	}

	_ = json.NewEncoder(w).Encode(authResponse)
}

func (s * authConfigServer) authPublicKey(w http.ResponseWriter, req *http.Request) {
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

	s.logger.DebugF("public key authentication request for user %s", authRequest.Username)

	authResponse := protocol.AuthResponse{
		Success: false,
	}
	if authRequest.Username == "foo" || authRequest.Username == "busybox" {
		authResponse.Success = true
	}

	_ = json.NewEncoder(w).Encode(authResponse)
}

func (s * authConfigServer) configHandler(w http.ResponseWriter, req *http.Request) {
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

	s.logger.DebugF("config request for user %s", configRequest.Username)

	if configRequest.Username == "busybox" {
		response.Config.DockerRun.Config.ContainerConfig.Image = "busybox"
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		s.logger.ErrorE(err)
	}
}

func main() {
	logConfig, err := log.NewConfig(log.LevelDebugString)
	if err != nil {
		panic(err)
	}
	logWriter := writer.NewJsonLogWriter()
	logger := log.NewLoggerPipeline(logConfig, logWriter)
	s := &authConfigServer{
		logger: logger,
	}
	http.HandleFunc("/pubkey", s.authPublicKey)
	http.HandleFunc("/password", s.authPassword)
	http.HandleFunc("/config", s.configHandler)
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		logger.CriticalE(err)
		os.Exit(1)
	}
}
