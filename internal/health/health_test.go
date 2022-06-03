package health_test

import (
    "go.containerssh.io/libcontainerssh/config"
    "go.containerssh.io/libcontainerssh/internal/health"
    "go.containerssh.io/libcontainerssh/log"
    service2 "go.containerssh.io/libcontainerssh/service"

	"testing"
	"time"
)

func TestOk(t *testing.T) {
	logger := log.NewTestLogger(t)
	cfg := config.HealthConfig{
		Enable: true,
		HTTPServerConfiguration: config.HTTPServerConfiguration{
			Listen: "127.0.0.1:23074",
		},
		Client: config.HTTPClientConfiguration{
			URL:            "http://127.0.0.1:23074",
			AllowRedirects: false,
			Timeout:        5 * time.Second,
		},
	}

	srv, err := health.New(cfg, logger)
	if err != nil {
		t.Fatal(err)
	}

	l := service2.NewLifecycle(srv)

	running := make(chan struct{})
	l.OnRunning(func(s service2.Service, l service2.Lifecycle) {
		running <- struct{}{}
	})

	go func() {
		_ = l.Run()
	}()

	<-running

	client, err := health.NewClient(cfg, logger)
	if err != nil {
		t.Fatal(err)
	}

	if client.Run() {
		t.Fatal("Health check did not fail, even though status is false.")
	}

	srv.ChangeStatus(true)
	if !client.Run() {
		t.Fatal("Health check failed, even though status is true.")
	}

	srv.ChangeStatus(false)
	if client.Run() {
		t.Fatal("Health check did not fail, even though status is false.")
	}
}
