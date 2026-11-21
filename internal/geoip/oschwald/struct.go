package oschwald

import backend "github.com/oschwald/geoip2-golang"

type geoIPLookupProvider struct {
	geo *backend.Reader
}
