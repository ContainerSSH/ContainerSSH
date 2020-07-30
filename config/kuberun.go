package config

import (
	v1 "k8s.io/api/core/v1"
	"time"
)

type KubeRunConfig struct {
	Connection KubeRunConnectionConfig `json:"connection" yaml:"connection" comment:"Kubernetes configuration options"`
	Pod        KubeRunPodConfig        `json:"pod" yaml:"pod" comment:"Container configuration"`
	Timeout    time.Duration           `json:"timeout" yaml:"timeout" comment:"Timeout for pod creation" default:"60s"`
}

type KubeRunConnectionConfig struct {
	Host    string `json:"host" yaml:"host" comment:"a host string, a host:port pair, or a URL to the base of the apiserver." default:"kubernetes.default.svc"`
	APIPath string `json:"path" yaml:"path" comment:"APIPath is a sub-path that points to an API root." default:"/api"`

	Username string `json:"username" yaml:"username" comment:"Username for basic authentication"`
	Password string `json:"password" yaml:"password" comment:"Password for basic authentication"`

	Insecure   bool   `json:"insecure" yaml:"insecure" comment:"Server should be accessed without verifying the TLS certificate." default:"false"`
	ServerName string `json:"serverName" yaml:"serverName" comment:"ServerName is passed to the server for SNI and is used in the client to check server certificates against."`

	CertFile string `json:"certFile" yaml:"certFile" comment:"File containing client certificate for TLS client certificate authentication." default:"/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"`
	KeyFile  string `json:"keyFile" yaml:"keyFile" comment:"File containing client key for TLS client certificate authentication"`
	CAFile   string `json:"cacertFile" yaml:"cacertFile" comment:"File containing trusted root certificates for the server"`

	CertData string `json:"cert" yaml:"cert" comment:"PEM-encoded certificate for TLS client certificate authentication"`
	KeyData  string `json:"key" yaml:"key" comment:"PEM-encoded client key for TLS client certificate authentication"`
	CAData   string `json:"cacert" yaml:"cacert" comment:"PEM-encoded trusted root certificates for the server"`

	BearerToken     string `json:"bearerToken" yaml:"bearerToken" comment:"Bearer (service token) authentication"`
	BearerTokenFile string `json:"bearerTokenFile" yaml:"bearerTokenFile" comment:"Path to a file containing a BearerToken. Set to /var/run/secrets/kubernetes.io/serviceaccount/token to use service token in a Kubernetes cluster."`

	QPS     float32       `json:"qps" yaml:"qps" comment:"QPS indicates the maximum QPS to the master from this client." default:"5"`
	Burst   int           `json:"burst" yaml:"burst" comment:"Maximum burst for throttle." default:"10"`
	Timeout time.Duration `json:"timeout" yaml:"timeout" comment:"Timeout"`
}

type KubeRunPodConfig struct {
	Namespace              string            `json:"namespace" yaml:"namespace" comment:"Namespace to run the pod in" default:"default"`
	ConsoleContainerNumber int               `json:"consoleContainerNumber" yaml:"consoleContainerNumber" comment:"Which container to attach the SSH connection to" default:"0"`
	Spec                   v1.PodSpec        `json:"podSpec" yaml:"podSpec" comment:"Pod specification to launch" default:"{\"containers\":[{\"name\":\"shell\",\"image\":\"janoszen/containerssh-image\"}]}"`
	Subsystems             map[string]string `json:"subsystems" yaml:"subsystems" comment:"Subsystem names and binaries map." default:"{\"sftp\":\"/usr/lib/openssh/sftp-server\"}"`
	DisableCommand         bool              `json:"disableCommand" yaml:"disableCommand" comment:"Disable command execution passed from SSH"`
}
