package server

import (
	"github.com/janoszen/containerssh/log"
	"github.com/janoszen/containerssh/metrics"
)

type MetricsServer struct {
	config    Config
	collector *metrics.MetricCollector
	logger    log.Logger
}
