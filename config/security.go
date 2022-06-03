package config

import (
	"fmt"
)

// SecurityConfig is the configuration structure for security settings.
type SecurityConfig struct {
	// DefaultMode sets the default execution policy for all other commands. It is recommended to set this to "disable"
	// if for restricted setups to avoid accidentally allowing new features coming in with version upgrades.
	DefaultMode SecurityExecutionPolicy `json:"defaultMode" yaml:"defaultMode"`

	// ForceCommand behaves similar to the OpenSSH ForceCommand option. When set this command overrides any command
	// requested by the client and executes this command instead. The original command supplied by the client will be
	// set in the `SSH_ORIGINAL_COMMAND` environment variable.
	//
	// Setting ForceCommand changes subsystem requests into exec requests for the backends.
	ForceCommand string `json:"forceCommand" yaml:"forceCommand"`

	// Env controls whether to allow or block setting environment variables.
	Env SecurityEnvConfig `json:"env" yaml:"env"`
	// Command controls whether to allow or block command ("exec") requests via SSh.
	Command CommandConfig `json:"command" yaml:"command"`
	// Shell controls whether to allow or block shell requests via SSh.
	Shell SecurityShellConfig `json:"shell" yaml:"shell"`
	// Subsystem controls whether to allow or block subsystem requests via SSH.
	Subsystem SubsystemConfig `json:"subsystem" yaml:"subsystem"`

	// Forwarding controls whether to allow or block connection, port or socket forwarding
	Forwarding ForwardingConfig `json:"forwarding" yaml:"forwarding"`

	// TTY controls how to treat TTY/PTY requests by clients.
	TTY SecurityTTYConfig `json:"tty" yaml:"tty"`

	// Signal configures how to handle signal requests to running programs.
	Signal SecuritySignalConfig `json:"signal" yaml:"signal"`

	// MaxSessions drives how many session channels can be open at the same time for a single network connection.
	// -1 means unlimited. It is strongly recommended to configure this to a sane value, e.g. 10.
	MaxSessions int `json:"maxSessions" yaml:"maxSessions" default:"-1"`
}

// Validate validates a shell configuration
func (c SecurityConfig) Validate() error {
	if err := c.DefaultMode.Validate(); err != nil {
		return wrap(err, "defaultMode")
	}
	if err := c.Env.Validate(); err != nil {
		return wrap(err, "env")
	}
	if err := c.Command.Validate(); err != nil {
		return wrap(err, "command")
	}
	if err := c.Shell.Validate(); err != nil {
		return wrap(err, "shell")
	}
	if err := c.Subsystem.Validate(); err != nil {
		return wrap(err, "subsystem")
	}
	if err := c.TTY.Validate(); err != nil {
		return wrap(err, "tty")
	}
	if err := c.Signal.Validate(); err != nil {
		return wrap(err, "signal")
	}
	if c.MaxSessions < -1 {
		return newError("maxSessions", "invalid maxSessions setting: %d", c.MaxSessions)
	}
	return nil
}

// SecurityEnvConfig configures setting environment variables.
type SecurityEnvConfig struct {
	// Mode configures how to treat environment variable requests by SSH clients.
	Mode SecurityExecutionPolicy `json:"mode" yaml:"mode" default:""`
	// Allow takes effect when Mode is ExecutionPolicyFilter and only allows the specified environment variables to be
	// set.
	Allow []string `json:"allow" yaml:"allow"`
	// Allow takes effect when Mode is not ExecutionPolicyDisable and disallows the specified environment variables to
	// be set.
	Deny []string `json:"deny" yaml:"deny"`
}

// Validate validates a shell configuration
func (e SecurityEnvConfig) Validate() error {
	if err := e.Mode.Validate(); err != nil {
		return wrap(err, "mode")
	}
	return nil
}

// CommandConfig controls command executions via SSH (exec requests).
type CommandConfig struct {
	// Mode configures how to treat command execution (exec) requests by SSH clients.
	Mode SecurityExecutionPolicy `json:"mode" yaml:"mode" default:""`
	// Allow takes effect when Mode is ExecutionPolicyFilter and only allows the specified commands to be
	// executed. Note that the match an exact match is performed to avoid shell injections, etc.
	Allow []string `json:"allow" yaml:"allow"`
}

// Validate validates a shell configuration
func (c CommandConfig) Validate() error {
	if err := c.Mode.Validate(); err != nil {
		return wrap(err, "mode")
	}
	return nil
}

// SecurityShellConfig controls shell executions via SSH.
type SecurityShellConfig struct {
	// Mode configures how to treat shell requests by SSH clients.
	Mode SecurityExecutionPolicy `json:"mode" yaml:"mode" default:""`
}

// Validate validates a shell configuration
func (s SecurityShellConfig) Validate() error {
	if err := s.Mode.Validate(); err != nil {
		return wrap(err, "mode")
	}
	return nil
}

// SubsystemConfig controls shell executions via SSH.
type SubsystemConfig struct {
	// Mode configures how to treat subsystem requests by SSH clients.
	Mode SecurityExecutionPolicy `json:"mode" yaml:"mode" default:""`
	// Allow takes effect when Mode is ExecutionPolicyFilter and only allows the specified subsystems to be
	// executed.
	Allow []string `json:"allow" yaml:"allow"`
	// Allow takes effect when Mode is not ExecutionPolicyDisable and disallows the specified subsystems to be executed.
	Deny []string `json:"deny" yaml:"deny"`
}

// Validate validates a subsystem configuration
func (s SubsystemConfig) Validate() error {
	if err := s.Mode.Validate(); err != nil {
		return wrap(err, "mode")
	}
	return nil
}

type ForwardingConfig struct {
	// ReverseForwardingMode configures how to treat reverse port forwarding requests from the container to the client.
	ReverseForwardingMode SecurityExecutionPolicy `json:"reverseForwardingMode" yaml:"reverseForwardingMode" default:"disable"`

	// ForwardingMode configures how to treat port forwarding requests from the client to the container. Enabling this setting also allows using ContainerSSH as a SOCKs proxy.
	ForwardingMode SecurityExecutionPolicy `json:"forwardingMode" yaml:"forwardingMode" default:"disable"`

	// SocketForwardingMode configures how to treat connection requests from the client to a unix socket in the container.
	SocketForwardingMode SecurityExecutionPolicy `json:"socketForwardingMode" yaml:"socketForwardingMode" default:"disable"`

	// SocketListenMode configures how to treat requests to listen for connections to a unix socket in the container.
	SocketListenMode SecurityExecutionPolicy `json:"socketListenMode" yaml:"socketListenMode" default:"disable"`

	// X11forwardingMode configures how to treat X11 forwarding requests from the container to the client
	X11ForwardingMode SecurityExecutionPolicy `json:"x11ForwardingMode" yaml:"x11ForwardingMode" default:"disable"`
}

func (f ForwardingConfig) Validate() error {
	if err := f.ReverseForwardingMode.Validate(); err != nil {
		return fmt.Errorf("invalid mode (%w)", err)
	}
	if err := f.ForwardingMode.Validate(); err != nil {
		return fmt.Errorf("invalid mode (%w)", err)
	}
	if err := f.SocketForwardingMode.Validate(); err != nil {
		return fmt.Errorf("invalid mode (%w)", err)
	}
	if err := f.SocketListenMode.Validate(); err != nil {
		return fmt.Errorf("invalid mode (%w)", err)
	}
	if err := f.X11ForwardingMode.Validate(); err != nil {
		return fmt.Errorf("invalid mode (%w)", err)
	}
	return nil
}

// SecurityTTYConfig controls how to treat TTY/PTY requests by clients.
type SecurityTTYConfig struct {
	// Mode configures how to treat TTY/PTY requests by SSH clients.
	Mode SecurityExecutionPolicy `json:"mode" yaml:"mode" default:""`
}

// Validate validates the TTY configuration
func (t SecurityTTYConfig) Validate() error {
	if err := t.Mode.Validate(); err != nil {
		return wrap(err, "mode")
	}
	return nil
}

// SecuritySignalConfig configures how signal forwarding requests are treated.
type SecuritySignalConfig struct {
	// Mode configures how to treat signal requests to running programs
	Mode SecurityExecutionPolicy `json:"mode" yaml:"mode" default:""`
	// Allow takes effect when Mode is ExecutionPolicyFilter and only allows the specified signals to be forwarded.
	Allow []string `json:"allow" yaml:"allow"`
	// Allow takes effect when Mode is not ExecutionPolicyDisable and disallows the specified signals to be forwarded.
	Deny []string `json:"deny" allow:"deny"`
}

// Validate validates the signal configuration
func (s SecuritySignalConfig) Validate() error {
	if err := s.Mode.Validate(); err != nil {
		return wrap(err, "mode")
	}
	return nil
}

// SecurityExecutionPolicy drives how to treat a certain request.
type SecurityExecutionPolicy string

const (
	// ExecutionPolicyUnconfigured falls back to the default mode. If unconfigured on a global level the default is to
	// "allow".
	ExecutionPolicyUnconfigured SecurityExecutionPolicy = ""

	// ExecutionPolicyEnable allows the execution of the specified method unless the specified option matches the
	// "deny" list.
	ExecutionPolicyEnable SecurityExecutionPolicy = "enable"

	// ExecutionPolicyFilter filters the execution against a specified allow list. If the allow list is empty or not
	// supported this ootion behaves like "disable".
	ExecutionPolicyFilter SecurityExecutionPolicy = "filter"

	// ExecutionPolicyDisable disables the specified method and does not take the allow or deny lists into account.
	ExecutionPolicyDisable SecurityExecutionPolicy = "disable"
)

// Validate validates the execution policy.
func (e SecurityExecutionPolicy) Validate() error {
	switch e {
	case ExecutionPolicyUnconfigured:
	case ExecutionPolicyEnable:
	case ExecutionPolicyFilter:
	case ExecutionPolicyDisable:
	default:
		return fmt.Errorf("invalid mode: %s", e)
	}
	return nil
}
