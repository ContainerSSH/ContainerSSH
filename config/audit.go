package config

// swagger:enum AuditStorage
type AuditStorage string

const (
	AuditStorage_None AuditStorage = "none"
	AuditStorage_File AuditStorage = "file"
	AuditStorage_S3   AuditStorage = "s3"
)

// swagger:enum AuditFormat
type AuditFormat string

const (
	AuditFormat_None      AuditFormat = "none"
	AuditFormat_Audit     AuditFormat = "audit"
	AuditFormat_Asciinema AuditFormat = "asciinema"
)

type AuditConfig struct {
	// Audit format
	Format AuditFormat `json:"format" yaml:"format" default:"none"`
	// Audit storage type
	Storage AuditStorage `json:"storage" yaml:"storage" default:"none"`
	// File audit logger configuration
	File AuditFileConfig `json:"file" yaml:"file"`
	// S3 configuration
	S3 AuditS3Config `json:"s3" yaml:"s3"`
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

type AuditS3Config struct {
	Local           string `json:"local" yaml:"local" default:"/var/lib/audit"`
	AccessKey       string `json:"accessKey" yaml:"accessKey"`
	SecretKey       string `json:"secretKey" yaml:"secretKey"`
	Bucket          string `json:"bucket" yaml:"bucket"`
	Region          string `json:"region" yaml:"region"`
	Endpoint        string `json:"endpoint" yaml:"endpoint"`
	CaCert          string `json:"cacert" yaml:"cacert"`
	ACL             string `json:"acl" yaml:"acl"`
	UploadPartSize  uint   `json:"uploadPartSize" yaml:"uploadPartSize"`
	ParallelUploads uint   `json:"parallelUploads" yaml:"parallelUploads"`
}
