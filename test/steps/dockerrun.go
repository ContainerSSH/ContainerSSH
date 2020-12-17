package steps

import (
	"fmt"

	"github.com/containerssh/configuration"
)

func (scenario *Scenario) ConfigureDocker(username string) error {
	if scenario.ConfigServer == nil {
		return fmt.Errorf("config server is not running")
	}

	appConfig := &configuration.AppConfig{}
	appConfig.Backend = "docker"

	scenario.ConfigServer.SetUserConfig(username, appConfig)
	return nil
}
