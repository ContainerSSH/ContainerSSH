package webhook

import (
    "go.containerssh.io/libcontainerssh/config"
    internalConfig "go.containerssh.io/libcontainerssh/internal/config"
    "go.containerssh.io/libcontainerssh/internal/geoip/dummy"
    "go.containerssh.io/libcontainerssh/internal/metrics"
    "go.containerssh.io/libcontainerssh/log"
)

// NewTestClient creates a configuration client, primarily for testing purposes.
func NewTestClient(clientConfig config.ClientConfig, logger log.Logger) (Client, error) {
	metricsCollector := metrics.New(dummy.New())
	return internalConfig.NewClient(clientConfig, logger, metricsCollector)
}
