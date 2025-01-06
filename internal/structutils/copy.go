package structutils

import "github.com/qdm12/reprint"

// Copy deep copies a struct pointer from source to destination.
func Copy(destination interface{}, source interface{}) error {
	return reprint.FromTo(source, destination)
}
