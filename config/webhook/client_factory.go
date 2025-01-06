package webhook

import (
    "go.containerssh.io/containerssh/config"
    internalConfig "go.containerssh.io/containerssh/internal/config"
    "go.containerssh.io/containerssh/internal/geoip/dummy"
    "go.containerssh.io/containerssh/internal/metrics"
    "go.containerssh.io/containerssh/log"
)

// NewTestClient creates a configuration client, primarily for testing purposes.
func NewTestClient(clientConfig config.ClientConfig, logger log.Logger) (Client, error) {
	metricsCollector := metrics.New(dummy.New())
	return internalConfig.NewClient(clientConfig, logger, metricsCollector)
}
