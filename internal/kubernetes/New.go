package kubernetes

import (
	"context"
	"net"
	"sync"

	"github.com/containerssh/containerssh/config"
	"github.com/containerssh/containerssh/internal/metrics"
	"github.com/containerssh/containerssh/internal/sshserver"
	"github.com/containerssh/containerssh/log"
	"github.com/containerssh/containerssh/message"
)

func New(
	client net.TCPAddr,
	connectionID string,
	config config.KubernetesConfig,
	logger log.Logger,
	backendRequestsMetric metrics.SimpleCounter,
	backendFailuresMetric metrics.SimpleCounter,
) (sshserver.NetworkConnectionHandler, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	if config.Pod.DisableAgent {
		logger.Warning(
			message.NewMessage(
				message.EKubernetesGuestAgentDisabled,
			"You are using the Kubernetes backend without the ContainerSSH Guest Agent. Several features will not work as expected. Please see https://containerssh.io/reference/image/ for details.",
		))
	}

	var clientFactory kubernetesClientFactory = &kubernetesClientFactoryImpl{
		backendRequestsMetric: backendRequestsMetric,
		backendFailuresMetric: backendFailuresMetric,
	}

	cli, err := clientFactory.get(
		context.Background(),
		config,
		logger,
	)
	if err != nil {
		return nil, err
	}

	return &networkHandler{
		mutex:        &sync.Mutex{},
		client:       client,
		connectionID: connectionID,
		config:       config,
		cli:          cli,
		pod:          nil,
		labels:       nil,
		logger:       logger,
		disconnected: false,
		done:         make(chan struct{}),
	}, nil
}
