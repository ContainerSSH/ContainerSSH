package metadata

// Value is a string value with extra data connected to it. The value is a string type.
//
// swagger:model MetadataValue
type Value struct {
	// Value contains the string for the current value.
	Value string `json:"value"`
	// Sensitive indicates that the metadata value contains sensitive data and should not be transmitted to
	// servers unnecessarily.
	Sensitive bool `json:"sensitive"`
}

// BinaryValue is a value containing binary data. The value is a binary data type.
//
// swagger:model BinaryMetadataValue
type BinaryValue struct {
	// Value contains the binary data for the current value.
	//
	// required: true
	// in: body
	// swagger:strfmt: byte
	Value []byte `json:"value"`
	// Sensitive indicates that the metadata value contains sensitive data and should not be transmitted to
	// servers unnecessarily.
	//
	// required: false
	// in: body
	Sensitive bool `json:"sensitive"`
}
