package kuberun

import "github.com/janoszen/containerssh/metrics"

var MetricNameBackendError = "kuberun_error"
var MetricBackendError = metrics.Metric{
	Name:   MetricNameBackendError,
	Labels: map[string]string{},
}
