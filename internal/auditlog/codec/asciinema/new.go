package asciinema

import (
	"github.com/containerssh/libcontainerssh/internal/geoip/geoipprovider"
	"github.com/containerssh/libcontainerssh/log"

	"github.com/containerssh/libcontainerssh/internal/auditlog/codec"
)

// NewEncoder Creates an encoder that writes in the Asciicast v2 format
// (see https://github.com/asciinema/asciinema/blob/develop/doc/asciicast-v2.md)
func NewEncoder(logger log.Logger, geoIPProvider geoipprovider.LookupProvider) codec.Encoder {
	return &encoder{
		logger:        logger,
		geoIPProvider: geoIPProvider,
	}
}
