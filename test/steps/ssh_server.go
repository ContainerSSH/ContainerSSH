package steps

import (
	"context"
	"fmt"

	"github.com/containerssh/configuration"
	"github.com/containerssh/service"
	"github.com/containerssh/structutils"

	"github.com/containerssh/containerssh"
)

func (scenario *Scenario) StartSSHServer() error {
	if scenario.Lifecycle != nil {
		return fmt.Errorf("SSH server is already running")
	}

	appConfig := configuration.AppConfig{}
	structutils.Defaults(&appConfig)
	if err := appConfig.SSH.GenerateHostKey(); err != nil {
		return err
	}
	appConfig.Auth.URL = "http://127.0.0.1:8080"
	appConfig.Auth.Password = true
	appConfig.ConfigServer.URL = "http://127.0.0.1:8081/config"
	appConfig.Metrics.Enable = true

	srv, err := containerssh.New(
		appConfig,
		scenario.LoggerFactory,
	)
	if err != nil {
		return err
	}

	lifecycle := service.NewLifecycle(srv)
	running := make(chan struct{})
	lifecycle.OnRunning(
		func(s service.Service, l service.Lifecycle) {
			running <- struct{}{}
		})
	go func() {
		_ = lifecycle.Run()
	}()
	<-running
	scenario.Lifecycle = lifecycle
	return nil
}

func (scenario *Scenario) StopSshServer() error {
	if scenario.Lifecycle == nil {
		return fmt.Errorf("SSH server is already stopped")
	}
	scenario.Lifecycle.Stop(context.Background())
	err := scenario.Lifecycle.Wait()
	if err != nil {
		return err
	}
	scenario.Lifecycle = nil
	return nil
}
