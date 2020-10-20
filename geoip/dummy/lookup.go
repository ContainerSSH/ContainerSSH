package dummy

import (
	"net"
)

func (g *GeoIPLookupProvider) Lookup(_ net.IP) (countryCode string) {
	return "XX"
}
