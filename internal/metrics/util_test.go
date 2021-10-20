package metrics_test

import (
	"net"
)

type geoIpLookupProvider struct {
	ips map[string]string
}

func (g *geoIpLookupProvider) Lookup(remoteAddr net.IP) (countryCode string) {
	if val, ok := g.ips[remoteAddr.String()]; ok {
		return val
	}
	return "XX"
}
