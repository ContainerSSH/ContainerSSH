package config

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
)

type DockerRunConfig struct {
	Host   string                   `json:"host" yaml:"host" comment:"Docker connect URL" default:"tcp://127.0.0.1:2375"`
	CaCert string                   `json:"cacert" yaml:"cacert" comment:"CA certificate for Docker connection embedded in the configuration in PEM format."`
	Cert   string                   `json:"cert" yaml:"cert" comment:"Client certificate in PEM format embedded in the configuration."`
	Key    string                   `json:"key" yaml:"key" comment:"Client key in PEM format embedded in the configuration."`
	Config DockerRunContainerConfig `json:"config" yaml:"config" comment:"Config configuration"`
}

type DockerRunContainerConfig struct {
	ContainerConfig container.Config         `json:"container" yaml:"container" comment:"Config configuration." default:"{\"Image\":\"janoszen/containerssh-image\"}"`
	HostConfig      container.HostConfig     `json:"host" yaml:"host" comment:"Host configuration"`
	NetworkConfig   network.NetworkingConfig `json:"network" yaml:"network" comment:"Network configuration"`
	ContainerName   string                   `json:"containername" yaml:"containername" comment:"Name for the container to be launched"`
	Subsystems      map[string]string        `json:"subsystems" yaml:"subsystems" comment:"Subsystem names and binaries map." default:"{\"sftp\":\"/usr/lib/openssh/sftp-server\"}"`
}
