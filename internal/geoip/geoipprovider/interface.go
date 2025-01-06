package geoipprovider

import "net"

// LookupProvider is an interface for all IP to country code providers.
type LookupProvider interface {
	// Lookup takes an IPv4 or IPv6 IP address and returns a country code. If the lookup fails it must return the code
	//        "XX".
	Lookup(remoteAddr net.IP) (countryCode string)
}
