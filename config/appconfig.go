package config

import (
	"fmt"

	"github.com/containerssh/libcontainerssh/internal/structutils"
	"github.com/containerssh/libcontainerssh/message"
)

// AppConfig is the root configuration object of ContainerSSH.
//goland:noinspection GoDeprecation
type AppConfig struct {
	// SSH contains the configuration for the SSH server.
	// swagger:ignore
	SSH SSHConfig `json:"ssh" yaml:"ssh"`
	// ConfigServer contains the settings for fetching the user-specific configuration.
	// swagger:ignore
	ConfigServer ClientConfig `json:"configserver" yaml:"configserver"`
	// Auth contains the configuration for user authentication.
	// swagger:ignore
	Auth AuthConfig `json:"auth" yaml:"auth"`
	// Log contains the configuration for the logging level.
	// swagger:ignore
	Log LogConfig `json:"log" yaml:"log"`
	// Metrics contains the configuration for the metrics server.
	// swagger:ignore
	Metrics MetricsConfig `json:"metrics" yaml:"metrics"`
	// GeoIP contains the configuration for the GeoIP lookups.
	// swagger:ignore
	GeoIP GeoIPConfig `json:"geoip" yaml:"geoip"`
	// Audit contains the configuration for audit logging and log upload.
	// swagger:ignore
	Audit AuditLogConfig `json:"audit" yaml:"audit"`
	// Health contains the configuration for the health check service.
	Health HealthConfig `json:"health" yaml:"health"`

	// Security contains the security restrictions on what can be executed. This option can be changed from the config
	// server.
	Security SecurityConfig `json:"security" yaml:"security"`
	// Backend defines which backend to use. This option can be changed from the config server.
	Backend Backend `json:"backend" yaml:"backend" default:"docker"`
	// Docker contains the configuration for the docker backend. This option can be changed from the config server.
	Docker DockerConfig `json:"docker,omitempty" yaml:"docker"`
	// DockerRun is a placeholder for the removed DockerRun backend. Filling this with anything but nil will yield a
	// validation error.
	DockerRun interface{} `json:"dockerrun,omitempty"`
	// Kubernetes contains the configuration for the kubernetes backend. This option can be changed from the config
	// server.
	Kubernetes KubernetesConfig `json:"kubernetes,omitempty" yaml:"kubernetes"`
	// KubeRun is a placeholder for the removed DockerRun backend. Filling this with anything but nil will yield a
	// validation error.
	KubeRun interface{} `json:"kuberun,omitempty"`
	// SSHProxy is the configuration for the SSH proxy backend, which forwards requests to a backing SSH server.
	SSHProxy SSHProxyConfig `json:"sshproxy,omitempty" yaml:"sshproxy"`
}

// Backend holds the possible values for backend selector.
type Backend string

// BackendValues returns all possible values for the Backend field.
func BackendValues() []Backend {
	return []Backend{
		BackendDocker,
		BackendKubernetes,
		BackendSSHProxy,
	}
}

// Validate checks if the configured backend is a valid one.
func (b Backend) Validate() error {
	switch b {
	case BackendDocker:
		fallthrough
	case BackendKubernetes:
		fallthrough
	case BackendSSHProxy:
		return nil
	case "":
		return fmt.Errorf("no backend configured")
	default:
		return fmt.Errorf("invalid backend: %s", b)
	}
}

const (
	BackendDocker     Backend = "docker"
	BackendKubernetes Backend = "kubernetes"
	BackendSSHProxy   Backend = "sshproxy"
)

// Default sets the default values for the configuration.
func (cfg *AppConfig) Default() {
	structutils.Defaults(cfg)
}

// Validate validates the configuration structure and returns an error if it is invalid.
//
// - dynamic enables the validation for dynamically configurable options.
func (cfg *AppConfig) Validate(dynamic bool) error {
	if cfg.DockerRun != nil {
		return message.NewMessage(
			message.EDockerRunRemoved,
			"You are using the `dockerrun` configuration which has been removed since ContainerSSH 0.5. Please switch to the Docker backend. For details and a migration guide see https://containerssh.io/deprecations/dockerrun/",
		)
	}
	if cfg.KubeRun != nil {
		return message.NewMessage(
			message.EKubeRunRemoved,
			"You are using the `kuberun` configuration which has been removed since ContainerSSH 0.5. Please switch to the Kubernetes backend. For details and a migration guide see https://containerssh.io/deprecations/kuberun/",
		)
	}

	queue := newValidationQueue()
	queue.add("ssh", &cfg.SSH)
	queue.add("configserver", &cfg.ConfigServer)
	queue.add("auth", &cfg.Auth)
	queue.add("log", &cfg.Log)
	queue.add("metrics", &cfg.Metrics)
	queue.add("geoip", &cfg.GeoIP)
	queue.add("audit", &cfg.Audit)
	queue.add("health", &cfg.Health)

	if cfg.ConfigServer.URL != "" && !dynamic {
		return queue.Validate()
	}
	queue.add("security", &cfg.Security)
	queue.add("backend", &cfg.Backend)
	switch cfg.Backend {
	case BackendDocker:
		queue.add("docker", &cfg.Docker)
	case BackendKubernetes:
		queue.add("kubernetes", &cfg.Kubernetes)
	case BackendSSHProxy:
		queue.add("sshproxy", &cfg.SSHProxy)
	}

	return queue.Validate()
}

type validatable interface {
	Validate() error
}

func newValidationQueue() *validationQueue {
	return &validationQueue{
		items: map[string]validatable{},
	}
}

type validationQueue struct {
	items map[string]validatable
}

func (v *validationQueue) add(name string, item validatable) {
	v.items[name] = item
}

func (v *validationQueue) Validate() error {
	for name, item := range v.items {
		if err := item.Validate(); err != nil {
			return wrap(err, name)
		}
	}
	return nil
}
