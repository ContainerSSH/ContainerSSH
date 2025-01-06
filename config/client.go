package config

//noinspection GoNameStartsWithPackageName
type ClientConfig struct {
	HTTPClientConfiguration `json:",inline" yaml:",inline"`

	// TransmitSensitiveMetadata enables sending sensitive metadata fields to the configuration webhook server.
	// If disabled, sensitive metadata fields are sanitized from the webhook request.
	TransmitSensitiveMetadata bool `json:"transmitSensitiveMetadata" yaml:"transmitSensitiveMetadata"`
}

// Validate validates the client configuration.
func (c *ClientConfig) Validate() error {
	if c.HTTPClientConfiguration.URL == "" {
		return nil
	}
	return c.HTTPClientConfiguration.Validate()
}
