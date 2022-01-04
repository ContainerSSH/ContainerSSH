// ContainerSSH Authentication and Configuration Server
//
// This OpenAPI document describes the API endpoints that are required for implementing an authentication
// and configuration server for ContainerSSH. (See https://github.com/containerssh/libcontainerssh for details.)
//
//     Schemes: http, https
//     Host: localhost
//     BasePath: /
//     Version: 0.5.0
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

	auth2 "github.com/containerssh/libcontainerssh/auth"
	"github.com/containerssh/libcontainerssh/config"
	configWebhook "github.com/containerssh/libcontainerssh/config/webhook"
	"github.com/containerssh/libcontainerssh/http"
	"github.com/containerssh/libcontainerssh/internal/auth"
	"github.com/containerssh/libcontainerssh/log"
	"github.com/containerssh/libcontainerssh/service"
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
	*auth2.ConnectionMetadata,
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
	*auth2.ConnectionMetadata,
	error,
) {
	if Username == "foo" || Username == "busybox" {
		return true, nil, nil
	}
	return false, nil, nil
}

// swagger:operation POST /authz Authentication authz
//
// Authorization
//
// ---
// parameters:
// - name: request
//   in: body
//   description: The authorization request
//   required: true
//   schema:
//     "$ref": "#/definitions/AuthorizationRequest"
// responses:
//   "200":
//     "$ref": "#/responses/AuthResponse"
func (a *authHandler) OnAuthorization(PrincipalUsername string, _ string,_ string, _ string) (
	bool,
	*auth2.ConnectionMetadata,
	error,
) {
	if PrincipalUsername == "foo" || PrincipalUsername == "busybox" {
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
//     "$ref": "#/definitions/Request"
// responses:
//   "200":
//     "$ref": "#/responses/Response"
func (c *configHandler) OnConfig(request config.Request) (config.AppConfig, error) {
	cfg := config.AppConfig{}

	if request.Username == "busybox" {
		cfg.Docker.Execution.Launch.ContainerConfig = &container.Config{}
		cfg.Docker.Execution.Launch.ContainerConfig.Image = "busybox"
		cfg.Docker.Execution.DisableAgent = true
		cfg.Docker.Execution.Mode = config.DockerExecutionModeSession
		cfg.Docker.Execution.ShellCommand = []string{"/bin/sh"}
	}

	return cfg, nil
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
	logger, err := log.NewLogger(config.LogConfig{
		Level:       config.LogLevelDebug,
		Format:      config.LogFormatLJSON,
		Destination: config.LogDestinationStdout,
	})
	if err != nil {
		panic(err)
	}
	authHTTPHandler := auth.NewHandler(&authHandler{}, logger)
	configHTTPHandler, err := configWebhook.NewHandler(&configHandler{}, logger)
	if err != nil {
		panic(err)
	}

	srv, err := http.NewServer(
		"authconfig",
		config.HTTPServerConfiguration{
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
