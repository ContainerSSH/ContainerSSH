package steps

import (
	"fmt"
	"github.com/janoszen/containerssh/test/auth"
)

func (scenario *Scenario) StartConfigServer() error {
	if scenario.AuthServer != nil {
		return fmt.Errorf("auth server is already running")
	}
	scenario.AuthServer = auth.NewMemoryAuthServer()
	return scenario.AuthServer.Start()
}

func (scenario *Scenario) StopConfigServer() error {
	if scenario.AuthServer == nil {
		return fmt.Errorf("auth server is not running")
	}
	err := scenario.AuthServer.Stop()
	if err != nil {
		return err
	}
	scenario.AuthServer = nil
	return nil
}
