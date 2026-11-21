package http

import (
	"fmt"
)

// RequestEncoding is the method by which the response is encoded.
type RequestEncoding string

// RequestEncodingDefault is the default encoding and encodes the body to JSON.
const RequestEncodingDefault = ""

// RequestEncodingJSON encodes the body to JSON.
const RequestEncodingJSON = "JSON"

// RequestEncodingWWWURLEncoded encodes the body via www-urlencoded.
const RequestEncodingWWWURLEncoded = "WWW-URLENCODED"

// Validate validates the RequestEncoding
func (r RequestEncoding) Validate() error {
	switch r {
	case RequestEncodingDefault:
		return nil
	case RequestEncodingJSON:
		return nil
	case RequestEncodingWWWURLEncoded:
		return nil
	default:
		return fmt.Errorf("invalid request encoding: %s", r)
	}
}
