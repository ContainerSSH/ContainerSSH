package geoip

import "net"

type LookupProvider interface {
	Lookup(remoteAddr net.IP) (countryCode string)
}
