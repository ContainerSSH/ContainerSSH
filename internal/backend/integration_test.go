package backend_test

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/containerssh/libcontainerssh/config"
	"github.com/containerssh/libcontainerssh/internal/backend"
	"github.com/containerssh/libcontainerssh/internal/geoip"
	"github.com/containerssh/libcontainerssh/internal/metrics"
	"github.com/containerssh/libcontainerssh/internal/sshserver"
	"github.com/containerssh/libcontainerssh/internal/structutils"
	"github.com/containerssh/libcontainerssh/internal/test"
	"github.com/containerssh/libcontainerssh/log"
	"github.com/containerssh/libcontainerssh/service"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ssh"
)

func TestSimpleContainerLaunch(t *testing.T) {
	authPort := test.GetNextPort(t, "auth server")
	sshPort := test.GetNextPort(t, "SSH")
	cfg := config.AppConfig{}
	structutils.Defaults(&cfg)
	cfg.Backend = "docker"
	cfg.Auth.URL = fmt.Sprintf("http://localhost:%d", authPort)
	err := cfg.SSH.GenerateHostKey()
	cfg.SSH.Listen = fmt.Sprintf("127.0.0.1:%d", sshPort)
	assert.NoError(t, err)

	backendLogger := log.NewTestLogger(t)
	geoIPLookupProvider, err := geoip.New(
		config.GeoIPConfig{
			Provider: config.GeoIPDummyProvider,
		},
	)
	assert.NoError(t, err)
	metricsCollector := metrics.New(
		geoIPLookupProvider,
	)
	b, err := backend.New(
		cfg,
		backendLogger,
		metricsCollector,
		sshserver.AuthResponseSuccess,
	)
	assert.NoError(t, err)

	sshServerLogger := log.NewTestLogger(t)
	sshServer, err := sshserver.New(cfg.SSH, b, sshServerLogger)
	assert.NoError(t, err)

	lifecycle := service.NewLifecycle(sshServer)
	running := make(chan struct{})
	lifecycle.OnRunning(
		func(s service.Service, l service.Lifecycle) {
			running <- struct{}{}
		})
	go func() {
		_ = lifecycle.Run()
	}()
	<-running

	processClientInteraction(t, cfg)

	lifecycle.Stop(context.Background())
	err = lifecycle.Wait()
	assert.NoError(t, err)
}

func processClientInteraction(t *testing.T, config config.AppConfig) {
	clientConfig := &ssh.ClientConfig{
		User: "foo",
		Auth: []ssh.AuthMethod{ssh.Password("bar")},
	}
	clientConfig.HostKeyCallback = func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		return nil
	}
	sshConnection, err := ssh.Dial("tcp", config.SSH.Listen, clientConfig)
	if !assert.NoError(t, err) {
		return
	}
	defer func() {
		if sshConnection != nil {
			_ = sshConnection.Close()
		}
	}()

	session, err := sshConnection.NewSession()
	assert.NoError(t, err)

	output, err := session.CombinedOutput("echo 'Hello world!'")
	assert.NoError(t, err)

	assert.NoError(t, sshConnection.Close())
	assert.EqualValues(t, []byte("Hello world!\n"), output)
}
