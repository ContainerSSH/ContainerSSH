package server

import (
	"github.com/containerssh/containerssh/log"
	"github.com/containerssh/containerssh/metrics"
)

func New(
	config Config,
	collector *metrics.MetricCollector,
	logger log.Logger,
) *MetricsServer {
	return &MetricsServer{
		config:    config,
		collector: collector,
		logger:    logger,
	}
}
