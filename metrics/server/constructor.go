package server

import (
	"github.com/janoszen/containerssh/log"
	"github.com/janoszen/containerssh/metrics"
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
