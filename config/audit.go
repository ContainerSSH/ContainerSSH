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
	// Intecept input/output. Very resource intensive
	InterceptIO bool `json:"interceptIo" yaml:"interceptIo" default:"false"`
	// File audit logger configuration
	File AuditFileConfig `json:"file" yaml:"file"`
}

type AuditFileConfig struct {
	Directory string `json:"directory" yaml:"directory" default:""`
}
