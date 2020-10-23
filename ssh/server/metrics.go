package server

import "github.com/containerssh/containerssh/metrics"

var MetricNameConnections = "containerssh_ssh_connections"
var MetricNameSuccessfulHandshake = "containerssh_ssh_handshake_successful"
var MetricNameFailedHandshake = "containerssh_ssh_handshake_failed"
var MetricNameCurrentConnections = "containerssh_ssh_current_connections"
var MetricConnections = metrics.Metric{
	Name:   MetricNameConnections,
	Labels: map[string]string{},
}
var MetricSuccessfulHandshake = metrics.Metric{
	Name:   MetricNameSuccessfulHandshake,
	Labels: map[string]string{},
}
var MetricFailedHandshake = metrics.Metric{
	Name:   MetricNameFailedHandshake,
	Labels: map[string]string{},
}
var MetricCurrentConnections = metrics.Metric{
	Name:   MetricNameCurrentConnections,
	Labels: map[string]string{},
}
