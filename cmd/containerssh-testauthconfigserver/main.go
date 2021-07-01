// ContainerSSH Authentication and Configuration Server
//
// This OpenAPI document describes the API endpoints that are required for implementing an authentication
// and configuration server for ContainerSSH. (See https://github.com/containerssh/containerssh for details.)
//
//     Schemes: http, https
//     Host: localhost
//     BasePath: /
<<<<<<< HEAD
//     Version: 0.4.1
=======
//     Version: 0.5.0
>>>>>>> 56f7de5 (Issue #56: Keyboard-interactive authentication)
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

<<<<<<< HEAD
	"github.com/containerssh/auth"
	"github.com/containerssh/configuration/v2"
	"github.com/containerssh/docker/v2"
=======
	"github.com/containerssh/auth/v2"
	"github.com/containerssh/configuration/v3"
>>>>>>> 56f7de5 (Issue #56: Keyboard-interactive authentication)
	"github.com/containerssh/http"
	"github.com/containerssh/log"
	"github.com/containerssh/service"
	"github.com/docker/docker/api/types/container"
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
	map[string]string,
	error,
) {
	if os.Getenv("CONTAINERSSH_ALLOW_ALL") == "1" ||
		Username == "foo" ||
		Username == "busybox" {
		return true, nil, nil
	}
	return false, nil, nil
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
	map[string]string,
	error,
) {
	if Username == "foo" || Username == "busybox" {
		return true, nil, nil
	}
	return false, nil, nil
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
		config.DockerRun.Config.ContainerConfig = &container.Config{}
		config.DockerRun.Config.ContainerConfig.Image = "busybox"

		config.Docker.Execution.Launch.ContainerConfig = &container.Config{}
		config.Docker.Execution.Launch.ContainerConfig.Image = "busybox"
		config.Docker.Execution.DisableAgent = true
		config.Docker.Execution.Mode = docker.ExecutionModeSession
		config.Docker.Execution.ShellCommand = []string{"/bin/sh"}
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
