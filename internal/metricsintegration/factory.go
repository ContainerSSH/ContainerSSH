package metricsintegration

import (
	"github.com/containerssh/libcontainerssh/config"
	"github.com/containerssh/libcontainerssh/internal/metrics"
	"github.com/containerssh/libcontainerssh/internal/sshserver"
)

func NewHandler(
	cfg config.MetricsConfig,
	metricsCollector metrics.Collector,
	backend sshserver.Handler,
) (sshserver.Handler, error) {
	if !cfg.Enable {
		return backend, nil
	}

	connectionsMetric := metricsCollector.MustCreateCounterGeo(
		MetricNameConnections,
		"connections",
		MetricHelpConnections,
	)
	currentConnectionsMetric := metricsCollector.MustCreateGaugeGeo(
		MetricNameCurrentConnections,
		"connections",
		MetricHelpCurrentConnections,
	)

	handshakeSuccessfulMetric := metricsCollector.MustCreateCounterGeo(
		MetricNameSuccessfulHandshake,
		"handshakes",
		MetricHelpSuccessfulHandshake,
	)
	handshakeFailedMetric := metricsCollector.MustCreateCounterGeo(
		MetricNameFailedHandshake,
		"handshakes",
		MetricHelpFailedHandshake,
	)

	return &metricsHandler{
		backend:                   backend,
		metricsCollector:          metricsCollector,
		connectionsMetric:         connectionsMetric,
		handshakeSuccessfulMetric: handshakeSuccessfulMetric,
		handshakeFailedMetric:     handshakeFailedMetric,
		currentConnectionsMetric:  currentConnectionsMetric,
	}, nil
}
