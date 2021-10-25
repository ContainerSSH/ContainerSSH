package backend_test

import (
	"context"
	"net"
	"sync"
	"testing"

	"github.com/containerssh/containerssh/config"
	"github.com/containerssh/containerssh/internal/backend"
	"github.com/containerssh/containerssh/internal/geoip"
	"github.com/containerssh/containerssh/internal/metrics"
	"github.com/containerssh/containerssh/internal/sshserver"
	"github.com/containerssh/containerssh/internal/structutils"
	"github.com/containerssh/containerssh/log"
	"github.com/containerssh/containerssh/service"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ssh"
)

func TestSimpleContainerLaunch(t *testing.T) {
	t.Parallel()

	lock := &sync.Mutex{}
	for _, backendName := range []string{"docker"} {
		t.Run("backend="+backendName, func(t *testing.T) {
			lock.Lock()
			defer lock.Unlock()
			cfg := config.AppConfig{}
			structutils.Defaults(&cfg)
			cfg.Backend = backendName
			cfg.Auth.URL = "http://localhost:8080"
			err := cfg.SSH.GenerateHostKey()
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
		})
	}
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
