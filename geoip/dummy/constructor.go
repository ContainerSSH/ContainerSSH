package dummy

import "github.com/janoszen/containerssh/geoip"

func New() geoip.LookupProvider {
	return &GeoIPLookupProvider{}
}
