package config

type AppConfig struct {
	Listen       string             `json:"listen" yaml:"listen" default:"0.0.0.0:2222" comment:"IP address and port to listen on"`
	Ssh          SshConfig          `json:"ssh" yaml:"ssh" comment:"SSH configuration"`
	ConfigServer ConfigServerConfig `json:"configserver" yaml:"configserver" comment:"Configuration server settings"`
	Auth         AuthConfig         `json:"auth" yaml:"auth" comment:"Authentication server configuration"`
	Backend      string             `json:"backend" yaml:"backend" default:"dockerrun" comment:"Backend module to use"`
	DockerRun    DockerRunConfig    `json:"dockerrun" yaml:"dockerrun" comment:"Docker configuration to use when the Docker run backend is used."`
}
