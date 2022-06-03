package geoip

import (
	"fmt"

    "go.containerssh.io/libcontainerssh/config"
    "go.containerssh.io/libcontainerssh/internal/geoip/dummy"
    "go.containerssh.io/libcontainerssh/internal/geoip/geoipprovider"
    "go.containerssh.io/libcontainerssh/internal/geoip/oschwald"
)

// New creates a new lookup provider based on the configuration.
func New(cfg config.GeoIPConfig) (geoipprovider.LookupProvider, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	switch cfg.Provider {
	case config.GeoIPDummyProvider:
		return dummy.New(), nil
	case config.GeoIPMaxMindProvider:
		return oschwald.New(cfg.GeoIP2File)
	default:
		return nil, fmt.Errorf("invalid provider: %s", cfg.Provider)
	}
}
