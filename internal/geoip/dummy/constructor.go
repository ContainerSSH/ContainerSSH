package dummy

import (
	"github.com/containerssh/containerssh/internal/geoip/geoipprovider"
)

// New creates a dummy provider that always responds with "XX"
func New() geoipprovider.LookupProvider {
	return &geoIPLookupProvider{}
}
