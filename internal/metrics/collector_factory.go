package metrics

import (
	"sync"

	"github.com/containerssh/containerssh/internal/geoip/geoipprovider"
)

// New creates the metric collector.
func New(geoIpLookupProvider geoipprovider.LookupProvider) Collector {
	return &collector{
		geoIpLookupProvider: geoIpLookupProvider,
		mutex:               &sync.Mutex{},
		metricsMap:          map[string]Metric{},
		metrics:             []Metric{},
		values:              map[string]*metricValue{},
	}
}
