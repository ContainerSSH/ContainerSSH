package dummy

import "github.com/containerssh/containerssh/geoip"

func New() geoip.LookupProvider {
	return &GeoIPLookupProvider{}
}
