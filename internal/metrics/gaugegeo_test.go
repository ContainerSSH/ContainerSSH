package metrics_test

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"

    "go.containerssh.io/libcontainerssh/internal/metrics"
)

func TestGaugeGeo(t *testing.T) {
	collector := metrics.New(&geoIpLookupProvider{})
	gauge, err := collector.CreateGaugeGeo("test", "seconds", "Hello world!")
	assert.Nil(t, err, "creating counter returned an error")

	gauge.Set(net.ParseIP("127.0.0.1"), 42)

	testMetrics := collector.GetMetric("test")
	assert.Equal(t, 1, len(testMetrics))
	assert.Equal(t, 1, len(testMetrics[0].Labels))
	assert.Equal(t, float64(42), testMetrics[0].Value)

	newGauge := gauge.WithLabels(metrics.Label("foo", "bar"))
	newGauge.Set(net.ParseIP("127.0.0.1"), 43)

	testMetrics = collector.GetMetric("test")
	assert.Equal(t, 2, len(testMetrics))
	assert.Equal(t, 1, len(testMetrics[0].Labels))
	assert.Equal(t, 2, len(testMetrics[1].Labels))
	assert.Equal(t, float64(42), testMetrics[0].Value)
	assert.Equal(t, float64(43), testMetrics[1].Value)
}
