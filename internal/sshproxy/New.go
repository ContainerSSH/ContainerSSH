package sshproxy

import (
	"net"
	"sync"

    "go.containerssh.io/libcontainerssh/config"
    "go.containerssh.io/libcontainerssh/internal/metrics"
    "go.containerssh.io/libcontainerssh/internal/sshserver"
    "go.containerssh.io/libcontainerssh/log"
)

func New(
	client net.TCPAddr,
	connectionID string,
	config config.SSHProxyConfig,
	logger log.Logger,
	backendRequestsMetric metrics.SimpleCounter,
	backendFailuresMetric metrics.SimpleCounter,
) (
	sshserver.NetworkConnectionHandler,
	error,
) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	privateKey, err := config.LoadPrivateKey()
	if err != nil {
		return nil, err
	}

	return &networkConnectionHandler{
		lock:                  &sync.Mutex{},
		wg:                    &sync.WaitGroup{},
		client:                client,
		connectionID:          connectionID,
		config:                config,
		logger:                logger.WithLabel("server", config.Server).WithLabel("port", config.Port),
		backendRequestsMetric: backendRequestsMetric,
		backendFailuresMetric: backendFailuresMetric,
		privateKey:            privateKey,
	}, nil
}
