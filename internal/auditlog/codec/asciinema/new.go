package asciinema

import (
    "go.containerssh.io/libcontainerssh/internal/auditlog/codec"
    "go.containerssh.io/libcontainerssh/internal/geoip/geoipprovider"
    "go.containerssh.io/libcontainerssh/log"
)

// NewEncoder Creates an encoder that writes in the Asciicast v2 format
// (see https://github.com/asciinema/asciinema/blob/develop/doc/asciicast-v2.md)
func NewEncoder(logger log.Logger, geoIPProvider geoipprovider.LookupProvider) codec.Encoder {
	return &encoder{
		logger:        logger,
		geoIPProvider: geoIPProvider,
	}
}
