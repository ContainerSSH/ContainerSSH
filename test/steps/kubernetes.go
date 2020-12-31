package steps

import (
	"fmt"

	"github.com/containerssh/configuration"
)

func (scenario *Scenario) ConfigureKubernetes(username string) error {
	if scenario.ConfigServer == nil {
		return fmt.Errorf("config server is not running")
	}

	appConfig := &configuration.AppConfig{}
	appConfig.Backend = "kubernetes"

	if err := appConfig.Kubernetes.SetConfigFromKubeConfig(); err != nil {
		return err
	}

	scenario.ConfigServer.SetUserConfig(username, appConfig)
	return nil
}
