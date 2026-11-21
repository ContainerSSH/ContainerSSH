package config

import (
	"fmt"
	"os"
)

// GeoIPProvider is a configuration option to select the GeoIP lookup provider.
type GeoIPProvider string

const (
	// GeoIPDummyProvider always returns the XX country code.
	GeoIPDummyProvider GeoIPProvider = "dummy"
	// GeoIPMaxMindProvider looks up IP addresses via the MaxMind GeoIP database.
	GeoIPMaxMindProvider GeoIPProvider = "maxmind"
)

func (p GeoIPProvider) Validate() error {
	switch p {
	case GeoIPDummyProvider:
	case GeoIPMaxMindProvider:
	default:
		return fmt.Errorf("invalid GeoIP provider: %s", p)
	}
	return nil
}

// GeoIPConfig is the structure configuring the GeoIP lookup process.
type GeoIPConfig struct {
	Provider   GeoIPProvider `yaml:"provider" json:"provider" default:"dummy"`
	GeoIP2File string        `yaml:"maxmind-geoip2-file" json:"maxmind-geoip2-file" default:"/var/lib/GeoIP/GeoIP2-Country.mmdb"`
}

// Validate checks the configuration.
func (config GeoIPConfig) Validate() error {
	if err := config.Provider.Validate(); err != nil {
		return wrap(err, "provider")
	}
	if config.Provider == GeoIPMaxMindProvider {
		stat, err := os.Stat(config.GeoIP2File)
		if err != nil {
			return wrapWithMessage(err, "maxmind-geoip2-file", "invalid MaxMind GeoIP2 file: %s", config.GeoIP2File)
		}
		if stat.IsDir() {
			return fmt.Errorf("invalid MaxMind GeoIP2 file: %s (is a directory)", config.GeoIP2File)
		}
	}
	return nil
}
