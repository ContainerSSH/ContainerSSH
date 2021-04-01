// ContainerSSH Authentication and Configuration Server
//
// This OpenAPI document describes the API endpoints that are required for implementing an authentication
// and configuration server for ContainerSSH. (See https://github.com/containerssh/containerssh for details.)
//
//     Schemes: http, https
//     Host: localhost
//     BasePath: /
//     Version: 0.4.0
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
	"context"
	goHttp "net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/containerssh/auth"
	"github.com/containerssh/configuration"
	"github.com/containerssh/http"
	"github.com/containerssh/log"
	"github.com/containerssh/service"
)

type authHandler struct {
}

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
func (a *authHandler) OnPassword(Username string, _ []byte, _ string, _ string) (
	bool,
	error,
) {
	if os.Getenv("CONTAINERSSH_ALLOW_ALL") == "1" ||
		Username == "foo" ||
		Username == "busybox" {
		return true, nil
	}
	return false, nil
}

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
func (a *authHandler) OnPubKey(Username string, _ string, _ string, _ string) (
	bool,
	error,
) {
	if Username == "foo" || Username == "busybox" {
		return true, nil
	}
	return false, nil
}

type configHandler struct {
}

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

func (c *configHandler) OnConfig(request configuration.ConfigRequest) (configuration.AppConfig, error) {
	config := configuration.AppConfig{}

	if request.Username == "busybox" {
		config.Docker.Execution.Launch.ContainerConfig.Image = "busybox"
		config.DockerRun.Config.ContainerConfig.Image = "busybox"
	}

	return config, nil
}

type handler struct {
	auth   goHttp.Handler
	config goHttp.Handler
}

func (h *handler) ServeHTTP(writer goHttp.ResponseWriter, request *goHttp.Request) {
	switch request.URL.Path {
	case "/password":
		fallthrough
	case "/pubkey":
		h.auth.ServeHTTP(writer, request)
	case "/config":
		h.config.ServeHTTP(writer, request)
	default:
		writer.WriteHeader(404)
	}
}

func main() {
	logger, err := log.NewLogger(log.Config{
		Level:       log.LevelDebug,
		Format:      log.FormatLJSON,
		Destination: log.DestinationStdout,
	})
	if err != nil {
		panic(err)
	}
	authHTTPHandler := auth.NewHandler(&authHandler{}, logger)
	configHTTPHandler, err := configuration.NewHandler(&configHandler{}, logger)
	if err != nil {
		panic(err)
	}

	srv, err := http.NewServer(
		"authconfig",
		http.ServerConfiguration{
			Listen: "0.0.0.0:8080",
		},
		&handler{
			auth:   authHTTPHandler,
			config: configHTTPHandler,
		},
		logger,
		func(s string) {

		},
	)
	if err != nil {
		panic(err)
	}

	running := make(chan struct{})
	stopped := make(chan struct{})
	lifecycle := service.NewLifecycle(srv)
	lifecycle.OnRunning(
		func(s service.Service, l service.Lifecycle) {
			println("Test Auth-Config Server is now running...")
			close(running)
		}).OnStopped(
		func(s service.Service, l service.Lifecycle) {
			close(stopped)
		})
	exitSignalList := []os.Signal{os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM}
	exitSignals := make(chan os.Signal, 1)
	signal.Notify(exitSignals, exitSignalList...)
	go func() {
		if err := lifecycle.Run(); err != nil {
			panic(err)
		}
	}()
	select {
	case <-running:
		if _, ok := <-exitSignals; ok {
			println("Stopping Test Auth-Config Server...")
			lifecycle.Stop(context.Background())
		}
	case <-stopped:
	}
}
