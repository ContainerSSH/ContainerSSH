package webhook_test

import (
	"context"
	"os"
	"time"

    "go.containerssh.io/libcontainerssh/config"
    "go.containerssh.io/libcontainerssh/config/webhook"
    "go.containerssh.io/libcontainerssh/log"
    "go.containerssh.io/libcontainerssh/service"
)

type myConfigReqHandler struct {
}

func (m *myConfigReqHandler) OnConfig(_ config.Request) (config.AppConfig, error) {
	return config.AppConfig{}, nil
}

// ExampleNewServer demonstrates how to set up a configuration webhook server.
func ExampleNewServer() {
	// Set up a logger
	logger := log.MustNewLogger(config.LogConfig{
		Level:       config.LogLevelWarning,
		Format:      config.LogFormatText,
		Destination: config.LogDestinationStdout,
		Stdout:      os.Stdout,
	})

	// Create a new config webhook server.
	srv, err := webhook.NewServer(
		config.HTTPServerConfiguration{
			Listen: "0.0.0.0:0",
		},
		&myConfigReqHandler{},
		logger,
	)
	if err != nil {
		// Handle error
		panic(err)
	}

	// Set up and run the web server service.
	lifecycle := service.NewLifecycle(srv)

	go func() {
		//Ignore error, handled later.
		_ = lifecycle.Run()
	}()

	// Sleep for 30 seconds as a test
	time.Sleep(30 * time.Second)

	// Set up a shutdown context to give a deadline for graceful shutdown.
	shutdownContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// Stop the server.
	lifecycle.Stop(shutdownContext)

	// Wait for the server to stop.
	lastError := lifecycle.Wait()
	if lastError != nil {
		// Server stopped abnormally.
		panic(lastError)
	}

	// Output:
}
