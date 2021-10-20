package webhook

import (
	"github.com/containerssh/containerssh/config"
	internalConfig "github.com/containerssh/containerssh/internal/config"
	"github.com/containerssh/containerssh/internal/geoip/dummy"
	"github.com/containerssh/containerssh/internal/metrics"
	"github.com/containerssh/containerssh/log"
)

// NewTestClient creates a configuration client, primarily for testing purposes.
func NewTestClient(clientConfig config.ClientConfig, logger log.Logger) (Client, error) {
	metricsCollector := metrics.New(dummy.New())
	return internalConfig.NewClient(clientConfig, logger, metricsCollector)
}


