package dummy

import (
	"net"
)

func (g *geoIPLookupProvider) Lookup(_ net.IP) (countryCode string) {
	return "XX"
}
