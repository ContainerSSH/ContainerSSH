// +build windows

package config

// DockerConnectionConfig configures how to connect to dockerd.
type DockerConnectionConfig struct {
	// Host is the docker connect URL
	Host string `json:"host" yaml:"host" default:"npipe:////./pipe/docker_engine"`
	// CaCert is the CA certificate for Docker connection embedded in the configuration in PEM format.
	CaCert string `json:"cacert" yaml:"cacert"`
	// Cert is the client certificate in PEM format embedded in the configuration.
	Cert string `json:"cert" yaml:"cert"`
	// Key is the client key in PEM format embedded in the configuration.
	Key string `json:"key" yaml:"key"`
}

//DockerRunConfig describes the old ContainerSSH 0.3 configuration format that can still be read and used.
//Deprecated: Switch to the more generic "docker" backend.
//goland:noinspection GoNameStartsWithPackageName,GoDeprecation
type DockerRunConfig struct {
	Host   string                   `json:"host" yaml:"host" comment:"Docker connect URL" default:"npipe:////./pipe/docker_engine"`
	CaCert string                   `json:"cacert" yaml:"cacert" comment:"CA certificate for Docker connection embedded in the configuration in PEM format."`
	Cert   string                   `json:"cert" yaml:"cert" comment:"Client certificate in PEM format embedded in the configuration."`
	Key    string                   `json:"key" yaml:"key" comment:"Client key in PEM format embedded in the configuration."`
	Config DockerRunContainerConfig `json:"config" yaml:"config" comment:"DockerConfig configuration"`
}
