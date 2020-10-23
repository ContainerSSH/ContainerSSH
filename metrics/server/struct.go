package server

import (
	"github.com/containerssh/containerssh/log"
	"github.com/containerssh/containerssh/metrics"
)

type MetricsServer struct {
	config    Config
	collector *metrics.MetricCollector
	logger    log.Logger
}
