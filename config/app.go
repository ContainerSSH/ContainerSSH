package config

import metricsServer "github.com/containerssh/containerssh/metrics/server"

// swagger:enum BackendName
type BackendName string

//goland:noinspection GoUnusedConst
const (
	BackendDockerRun BackendName = "dockerrun"
	BackendKubeRun   BackendName = "kuberun"
)

type AppConfig struct {
	// swagger:ignore
	Listen string `json:"listen" yaml:"listen" default:"0.0.0.0:2222" comment:"IP address and port to listen on"`
	// swagger:ignore
	Ssh SshConfig `json:"ssh" yaml:"ssh" comment:"SSH configuration"`
	// swagger:ignore
	ConfigServer ConfigServerConfig `json:"configserver" yaml:"configserver" comment:"Configuration server settings"`
	// swagger:ignore
	Auth AuthConfig `json:"auth" yaml:"auth" comment:"Authentication server configuration"`
	// Backend to use.
	Backend BackendName `json:"backend" yaml:"backend" default:"dockerrun" comment:"Backend module to use"`
	// Configuration for the dockerrun backend
	DockerRun DockerRunConfig `json:"dockerrun" yaml:"dockerrun" comment:"Docker configuration to use when the Docker run backend is used."`
	// Configuration for the kuberun backend
	KubeRun KubeRunConfig `json:"kuberun" yaml:"kuberun" comment:"Kubernetes configuration to use when the Kubernetes run backend is used."`
	// Logging configuration
	Log AppLogConfig `json:"log" yaml:"log" comment:"Log configuration"`
	// Metrics configuration
	Metrics metricsServer.Config `json:"metrics" yaml:"metrics" comment:"Metrics configuration."`
	// GeoIP configuration
	GeoIP GeoIPConfig `json:"geoip" yaml:"geoip" comment:"GeoIP database"`
	// Audit configuration
	Audit AuditConfig `json:"audit" yaml:"audit" comment:"Audit configuration"`
}
