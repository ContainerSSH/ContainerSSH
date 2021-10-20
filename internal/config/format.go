package config

import (
	"fmt"
)

// Format describes the format of the file being read.
type Format string

const (
	// FormatJSON reads/writes in JSON format
	FormatJSON Format = "json"
	// FormatYAML reads/writes in YAML format
	FormatYAML Format = "yaml"
)

// Validate validates the given format.
func (f Format) Validate() error {
	switch f {
	case FormatJSON:
		fallthrough
	case FormatYAML:
		return nil
	default:
		return fmt.Errorf("invalid format: %s", f)
	}
}
