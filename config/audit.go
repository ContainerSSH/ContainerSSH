package config

// swagger:enum AuditPluginType
type AuditPluginType string

const (
	AuditPluginType_None AuditPluginType = "none"
	AuditPluginType_Log  AuditPluginType = "log"
	AuditPluginType_File AuditPluginType = "file"
)

type AuditConfig struct {
	// Audit plugin type
	Plugin AuditPluginType `json:"plugin" yaml:"plugin" default:"none"`
	// File audit logger configuration
	File AuditFileConfig `json:"file" yaml:"file"`
	// What to intercept during the connection
	Intercept AuditInterceptConfig `json:"intercept" yaml:"intercept"`
}

type AuditInterceptConfig struct {
	Stdin     bool `json:"stdin" yaml:"stdin" default:"false"`
	Stdout    bool `json:"stdout" yaml:"stdout" default:"false"`
	Stderr    bool `json:"stderr" yaml:"stderr" default:"false"`
	Passwords bool `json:"passwords" yaml:"passwords" default:"false"`
}

type AuditFileConfig struct {
	Directory string `json:"directory" yaml:"directory" default:"/var/log/audit"`
}
