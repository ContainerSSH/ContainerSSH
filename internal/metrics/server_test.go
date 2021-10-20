package metrics_test

import (
	"context"
	"io/ioutil"
	"net"
	goHttp "net/http"
	"strings"
	"testing"

	"github.com/containerssh/http"
	"github.com/containerssh/containerssh/log"
	"github.com/containerssh/service"
	"github.com/stretchr/testify/assert"

	"github.com/containerssh/metrics"
)

func TestFetchMetrics(t *testing.T) {
	geoip := &geoIpLookupProvider{}
	logger := log.NewTestLogger(t)
	m := metrics.New(geoip)
	counter, err := m.CreateCounter("test", "pc", "Test metric")
	assert.Nil(t, err)

	s, err := metrics.NewServer(
		metrics.Config{
			ServerConfiguration: http.ServerConfiguration{
				Listen: "127.0.0.1:8080",
			},
			Enable: true,
			Path:   "/metrics",
		},
		m,
		logger,
	)
	assert.Nil(t, err)
	lifecycle := service.NewLifecycle(s)
	ready := make(chan struct{}, 1)
	lifecycle.OnRunning(func(s service.Service, l service.Lifecycle) {
		ready <- struct{}{}
	})
	go func() {
		err := lifecycle.Run()
		assert.Nil(t, err)
	}()
	defer func() {
		lifecycle.Stop(context.Background())
	}()
	<-ready
	bodyLines := callServer(t)
	assert.Contains(t, bodyLines, "# HELP test Test metric")
	assert.Contains(t, bodyLines, "# TYPE test counter")
	assert.Contains(t, bodyLines, "# UNIT test pc")
	assert.Equal(t, "# EOF", bodyLines[len(bodyLines)-2])

	counter.Increment()
	bodyLines = callServer(t)
	// TODO this relies on an implementation detail and should be done nicer. Fix it the first time it breaks.
	assert.Contains(t, bodyLines, "test 1.000000")

	counter.Increment()
	bodyLines = callServer(t)
	// TODO this relies on an implementation detail and should be done nicer. Fix it the first time it breaks.
	assert.Contains(t, bodyLines, "test 2.000000")

	counter2, err := m.CreateCounterGeo("test2", "pc", "Test metric 2")
	assert.Nil(t, err)
	counter2.Increment(net.ParseIP("127.0.0.1"))
	bodyLines = callServer(t)
	assert.Contains(t, bodyLines, "test2{country=\"XX\"} 1.000000")
}

func callServer(t *testing.T) []string {
	metricsResult, err := goHttp.Get("http://127.0.0.1:8080/metrics")
	assert.Nil(t, err)
	bodyBytes, err := ioutil.ReadAll(metricsResult.Body)
	assert.Nil(t, err)
	assert.Nil(t, metricsResult.Body.Close())
	bodyLines := strings.Split(string(bodyBytes), "\n")
	return bodyLines
}
