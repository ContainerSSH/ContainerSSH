package config

import (
	"fmt"

	"github.com/containerssh/libcontainerssh/internal/structutils"
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
	Backend string `json:"backend" yaml:"backend" default:"docker"`
	// Docker contains the configuration for the docker backend. This option can be changed from the config server.
	Docker DockerConfig `json:"docker,omitempty" yaml:"docker"`
	// Kubernetes contains the configuration for the kubernetes backend. This option can be changed from the config
	// server.
	Kubernetes KubernetesConfig `json:"kubernetes,omitempty" yaml:"kubernetes"`
	// SSHProxy is the configuration for the SSH proxy backend, which forwards requests to a backing SSH server.
	SSHProxy SSHProxyConfig `json:"sshproxy,omitempty" yaml:"sshproxy"`
}

// Default sets the default values for the configuration.
func (cfg *AppConfig) Default() {
	structutils.Defaults(cfg)
}

// Validate validates the configuration structure and returns an error if it is invalid.
//
// - dynamic enables the validation for dynamically configurable options.
func (cfg *AppConfig) Validate(dynamic bool) error {
	queue := newValidationQueue()
	queue.add("SSH", &cfg.SSH)
	queue.add("config server", &cfg.ConfigServer)
	queue.add("authentication", &cfg.Auth)
	queue.add("logging", &cfg.Log)
	queue.add("metrics", &cfg.Metrics)
	queue.add("GeoIP", &cfg.GeoIP)
	queue.add("audit log", &cfg.Audit)
	queue.add("health", &cfg.Health)

	if cfg.ConfigServer.URL != "" && !dynamic {
		return queue.Validate()
	}
	queue.add("security configuration", &cfg.Security)
	switch cfg.Backend {
	case "docker":
		queue.add("Docker", &cfg.Docker)
	case "kubernetes":
		queue.add("Kubernetes", &cfg.Kubernetes)
	case "sshproxy":
		queue.add("SSH proxy", &cfg.SSHProxy)
	case "":
		return fmt.Errorf("no backend configured")
	default:
		return fmt.Errorf("invalid backend: %s", cfg.Backend)
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
			return fmt.Errorf("invalid %s configuration (%w)", name, err)
		}
	}
	return nil
}
