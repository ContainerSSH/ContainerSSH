package server

type Config struct {
	Enable bool   `yaml:"enable" json:"enable" comment:"Enable metrics server." default:"false"`
	Listen string `yaml:"listen" json:"listen" comment:"Listen on this address." default:"0.0.0.0:9100"`
	Path   string `yaml:"path" json:"path" comment:"Path to run the Metrics endpoint on." default:"/metrics"`
}
