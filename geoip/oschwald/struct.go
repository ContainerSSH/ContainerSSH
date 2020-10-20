package oschwald

import backend "github.com/oschwald/geoip2-golang"

type GeoIPLookupProvider struct {
	geo *backend.Reader
}
