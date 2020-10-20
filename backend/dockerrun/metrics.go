package dockerrun

import "github.com/janoszen/containerssh/metrics"

var MetricNameBackendError = "dockerrun_error"
var MetricBackendError = metrics.Metric{
	Name:   MetricNameBackendError,
	Labels: map[string]string{},
}
