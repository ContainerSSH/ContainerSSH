package config

// HealthConfig is the configuration for the health service.
type HealthConfig struct {
	Enable                  bool `json:"enable" yaml:"enable"`
	HTTPServerConfiguration `json:",inline" yaml:",inline" default:"{\"listen\":\"0.0.0.0:7000\"}"`
	Client                  HTTPClientConfiguration `json:"client" yaml:"client" default:"{\"url\":\"http://127.0.0.1:7000/\"}"`
}

func (c HealthConfig) Validate() error {
	if !c.Enable {
		return nil
	}
	if err := c.HTTPServerConfiguration.Validate(); err != nil {
		return err
	}
	if err := c.Client.Validate(); err != nil {
		return wrap(err, "client")
	}
	return nil
}
