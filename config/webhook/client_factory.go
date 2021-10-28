package webhook

import (
	"github.com/containerssh/libcontainerssh/config"
	internalConfig "github.com/containerssh/libcontainerssh/internal/config"
	"github.com/containerssh/libcontainerssh/internal/geoip/dummy"
	"github.com/containerssh/libcontainerssh/internal/metrics"
	"github.com/containerssh/libcontainerssh/log"
)

// NewTestClient creates a configuration client, primarily for testing purposes.
func NewTestClient(clientConfig config.ClientConfig, logger log.Logger) (Client, error) {
	metricsCollector := metrics.New(dummy.New())
	return internalConfig.NewClient(clientConfig, logger, metricsCollector)
}
