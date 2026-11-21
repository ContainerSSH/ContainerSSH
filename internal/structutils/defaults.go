package structutils

import "github.com/creasty/defaults"

// Defaults fills a struct pointer with default values from the "default" tag. It only fills public fields.
//
// - Scalars can be provided directly
// - maps, structs, etc. can be provided in JSON format
// - time.Duration can be provided in text format (e.g. 60s)
func Defaults(data interface{}) {
	defaults.MustSet(data)
}

// DefaultsProvider is a struct that has a custom default setter.
type DefaultsProvider interface {
	// SetDefaults sets the default values on the current implementation.
	SetDefaults()
}
