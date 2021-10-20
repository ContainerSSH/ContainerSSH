package metrics

import (
	"net/http"
)

// NewHandler creates a new HTTP handler that outputs the collected metrics to the requesting client in the
//            Prometheus/OpenMetrics format.
func NewHandler(
	path string,
	collector Collector,
) http.Handler {
	if collector == nil {
		panic("the collector passed to server.NewHandler() must not be nil")
	}
	return &metricsHandler{
		path:      path,
		collector: collector,
	}
}

type metricsHandler struct {
	collector Collector
	path      string
}

func (h *metricsHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if h.path != request.URL.Path {
		writer.WriteHeader(404)
		return
	}
	writer.Header().Set("Content-Type", "application/openmetrics-text; version=1.0.0; charset=utf-8")
	writer.WriteHeader(200)
	_, _ = writer.Write([]byte(h.collector.String()))
}
