package metrics_test

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"

    "go.containerssh.io/libcontainerssh/internal/metrics"
)

// TestCounterGeo tests the functionality of the Geo counter
func TestCounterGeo(t *testing.T) {
	geoip := &geoIpLookupProvider{ips: map[string]string{
		"127.0.0.1": "LO",
	}}
	collector := metrics.New(geoip)
	counter, err := collector.CreateCounterGeo("test", "seconds", "Hello world!")
	assert.Nil(t, err, "creating counter returned an error")

	m := collector.ListMetrics()
	assert.Equal(t, 1, len(m))
	assert.Equal(t, "test", m[0].Name)
	assert.Equal(t, "seconds", m[0].Unit)
	assert.Equal(t, "Hello world!", m[0].Help)
	assert.Equal(t, metrics.MetricTypeCounter, m[0].Type)
	assert.Equal(t, 0, len(collector.GetMetric("test")))

	counter.Increment(net.ParseIP("127.0.0.1"))
	metric := collector.GetMetric("test")
	assert.Equal(t, 1, len(metric))
	assert.Equal(t, "test", metric[0].Name)
	assert.Equal(t, float64(1), metric[0].Value)
	assert.Equal(t, map[string]string{"country": "LO"}, metric[0].Labels)

	counter.Increment(net.ParseIP("127.0.0.2"))
	metric = collector.GetMetric("test")
	collectedMetrics := make([]string, len(metric))
	for i, m := range metric {
		collectedMetrics[i] = m.CombinedName()
	}
	assert.Contains(t, collectedMetrics, "test{country=\"LO\"}")
	assert.Contains(t, collectedMetrics, "test{country=\"XX\"}")
	assert.Equal(t, 2, len(metric))
	assert.Equal(t, float64(1), metric[0].Value)
	assert.Equal(t, float64(1), metric[1].Value)

	counter.Increment(net.ParseIP("127.0.0.2"), metrics.Label("foo", "bar"))
	metric = collector.GetMetric("test")
	for _, m := range metric {
		if m.CombinedName() == "test{country=\"XX\",foo=\"bar\"}" {
			assert.Equal(t, float64(1), m.Value)
		}
	}
}

// TestCounter tests the functionality of counters
func TestGeoCounterLabel(t *testing.T) {
	collector := metrics.New(&geoIpLookupProvider{})
	counter, err := collector.CreateCounterGeo("test", "seconds", "Hello world!")
	assert.Nil(t, err, "creating counter returned an error")
	counter.Increment(net.ParseIP("127.0.0.2"))
	newCounter := counter.WithLabels(metrics.Label("foo", "bar"))
	newCounter.Increment(net.ParseIP("127.0.0.2"), metrics.Label("baz", "bar"))

	metric := collector.GetMetric("test")
	assert.Equal(t, 2, len(metric))

	assert.Equal(t, "test", metric[0].Name)
	assert.Equal(t, float64(1), metric[0].Value)
	assert.Equal(t, 1, len(metric[0].Labels))

	assert.Equal(t, "test", metric[1].Name)
	assert.Equal(t, float64(1), metric[1].Value)
	assert.Equal(t, 3, len(metric[1].Labels))
}
