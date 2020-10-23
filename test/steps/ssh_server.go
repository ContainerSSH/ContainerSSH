package steps

import (
	"fmt"
	authClient "github.com/containerssh/containerssh/auth"
	"github.com/containerssh/containerssh/config"
	configClient "github.com/containerssh/containerssh/config/client"
	"github.com/containerssh/containerssh/geoip/dummy"
	"github.com/containerssh/containerssh/metrics"
	"github.com/containerssh/containerssh/test/ssh"
	"net"
	"time"
)

func (scenario *Scenario) StartSshServer() error {
	if scenario.SshServer != nil {
		return fmt.Errorf("SSH server is already running")
	}
	metric := metrics.New(dummy.New())
	ac, err := authClient.NewHttpAuthClient(
		config.AuthConfig{
			Url: "http://127.0.0.1:8080",
		},
		scenario.Logger,
		metric,
	)
	if err != nil {
		return err
	}

	configClientInstance, err := configClient.NewHttpConfigClient(
		config.ConfigServerConfig{
			Url: "http://127.0.0.1:8081/config",
		},
		scenario.Logger,
		metric,
	)
	if err != nil {
		return err
	}

	scenario.SshServer = ssh.NewServer(
		scenario.Logger,
		scenario.LogWriter,
		ac,
		configClientInstance,
	)
	err = scenario.SshServer.Start()
	if err != nil {
		return err
	}
	tries := 0
	for {
		tcp, err := net.Dial("tcp", "127.0.0.1:2222")
		if err == nil {
			_ = tcp.Close()
			break
		}
		tries = tries + 1
		if tries > 100 {
			_ = scenario.SshServer.Stop()
			scenario.SshServer = nil
			return fmt.Errorf("failed to start SSH server")
		}
		time.Sleep(time.Millisecond * 100)
	}
	return nil
}

func (scenario *Scenario) StopSshServer() error {
	if scenario.SshServer == nil {
		return fmt.Errorf("SSH server is already stopped")
	}
	err := scenario.SshServer.Stop()
	if err != nil {
		return err
	}
	scenario.SshServer = nil
	return nil
}
