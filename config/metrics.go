package config

type MetricsConfig struct {
	HTTPServerConfiguration `json:",inline" yaml:",inline" default:"{\"listen\":\"0.0.0.0:9100\"}"`

	Enable bool   `yaml:"enable" json:"enable" comment:"Enable metrics server." default:"false"`
	Path   string `yaml:"path" json:"path" comment:"Path to run the Metrics endpoint on." default:"/metrics"`
}

// Validate validates the configuration.
func (c MetricsConfig) Validate() error {
	if !c.Enable {
		return nil
	}
	if c.Path == "" {
		return newError("path", "metrics path cannot be empty")
	}
	return c.HTTPServerConfiguration.Validate()
}
