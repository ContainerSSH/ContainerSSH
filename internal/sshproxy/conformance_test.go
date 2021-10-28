package sshproxy_test

import (
	"net"
	"testing"

	"github.com/containerssh/libcontainerssh/config"
	"github.com/containerssh/libcontainerssh/internal/geoip/dummy"
	"github.com/containerssh/libcontainerssh/internal/metrics"
	"github.com/containerssh/libcontainerssh/internal/sshproxy"
	"github.com/containerssh/libcontainerssh/internal/sshserver"
	"github.com/containerssh/libcontainerssh/internal/structutils"
	"github.com/containerssh/libcontainerssh/internal/test"
	"github.com/containerssh/libcontainerssh/log"
)

func TestConformance(t *testing.T) {
	var factories = map[string]func(t *testing.T, logger log.Logger) (sshserver.NetworkConnectionHandler, error){
		"sshproxy": func(t *testing.T, logger log.Logger) (sshserver.NetworkConnectionHandler, error) {
			var err error
			sshServer := test.SSH(t)
			connectionID := sshserver.GenerateConnectionID()
			cfg := config.SSHProxyConfig{}
			structutils.Defaults(&cfg)
			cfg.Server = sshServer.Host()
			cfg.Port = uint16(sshServer.Port())
			cfg.Username = sshServer.Username()
			cfg.Password = sshServer.Password()
			cfg.AllowedHostKeyFingerprints = []string{
				sshServer.FingerprintSHA256(),
			}
			cfg.HostKeyAlgorithms, err = config.SSHKeyAlgoListFromStringList(sshServer.HostKeyAlgorithms())
			if err != nil {
				t.Fatalf("invalid SSH host key algorithm list (%v)", err)
			}
			collector := metrics.New(dummy.New())
			return sshproxy.New(
				net.TCPAddr{
					IP:   net.ParseIP("127.0.0.1"),
					Port: 2222,
					Zone: "",
				},
				connectionID,
				cfg,
				logger,
				collector.MustCreateCounter("backend_requests", "", ""),
				collector.MustCreateCounter("backend_failures", "", ""),
			)
		},
	}

	sshserver.RunConformanceTests(t, factories)
}
