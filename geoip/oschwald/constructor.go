package oschwald

import (
	"github.com/containerssh/containerssh/geoip"
	backend "github.com/oschwald/geoip2-golang"
)

func New(geoIpFile string) (geoip.LookupProvider, error) {
	geo, err := backend.Open(geoIpFile)
	if err != nil {
		return nil, err
	}
	return &GeoIPLookupProvider{
		geo: geo,
	}, nil
}
