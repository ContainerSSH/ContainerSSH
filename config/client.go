package config

//noinspection GoNameStartsWithPackageName
type ClientConfig struct {
	HTTPClientConfiguration `json:",inline" yaml:",inline"`
}

// Validate validates the client configuration.
func (c *ClientConfig) Validate() error {
	if c.HTTPClientConfiguration.URL == "" {
		return nil
	}
	return c.HTTPClientConfiguration.Validate()
}
