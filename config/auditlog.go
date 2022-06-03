package config

import (
	"fmt"
	"os"
)

// AuditLogFormat describes the audit log format in use.
type AuditLogFormat string

const (
	// AuditLogFormatNone signals that no audit logging should take place.
	AuditLogFormatNone AuditLogFormat = "none"
	// AuditLogFormatBinary signals that audit logging should take place in CBOR+GZIP format
	//              (see https://containerssh.github.io/advanced/audit/format/ )
	AuditLogFormatBinary AuditLogFormat = "binary"
	// AuditLogFormatAsciinema signals that audit logging should take place in Asciicast v2 format
	//                 (see https://github.com/asciinema/asciinema/blob/develop/doc/asciicast-v2.md )
	AuditLogFormatAsciinema AuditLogFormat = "asciinema"
)

// Validate checks the format.
func (f AuditLogFormat) Validate() error {
	switch f {
	case AuditLogFormatBinary:
	case AuditLogFormatAsciinema:
	case AuditLogFormatNone:
	default:
		return fmt.Errorf("invalid audit log format: %s", f)
	}
	return nil
}

// AuditLogStorage describes the storage backend to use.
type AuditLogStorage string

const (
	// AuditLogStorageNone signals that no storage should be used.
	AuditLogStorageNone AuditLogStorage = "none"
	// AuditLogStorageFile signals that audit logs should be stored in a local directory.
	AuditLogStorageFile AuditLogStorage = "file"
	// AuditLogStorageS3 signals that audit logs should be stored in an S3-compatible object storage.
	AuditLogStorageS3 AuditLogStorage = "s3"
)

// Validate checks the storage.
func (s AuditLogStorage) Validate() error {
	switch s {
	case AuditLogStorageNone:
	case AuditLogStorageFile:
	case AuditLogStorageS3:
	default:
		return fmt.Errorf("invalid audit log storage: %s", s)
	}
	return nil
}

// AuditLogConfig is the configuration structure for audit logging.
type AuditLogConfig struct {
	// Enable turns on audit logging.
	Enable bool `json:"enable" yaml:"enable" default:"false"`
	// Format audit format
	Format AuditLogFormat `json:"format" yaml:"format" default:"none"`
	// Storage audit storage type
	Storage AuditLogStorage `json:"storage" yaml:"storage" default:"none"`
	// File audit logger configuration
	File AuditLogFileConfig `json:"file" yaml:"file"`
	// S3 configuration
	S3 AuditLogS3Config `json:"s3" yaml:"s3"`
	// Intercept configures what should be intercepted
	Intercept AuditLogInterceptConfig `json:"intercept" yaml:"intercept"`
}

// AuditLogInterceptConfig configures what should be intercepted by the auditing facility.
type AuditLogInterceptConfig struct {
	// Stdin signals that the standard input from the user should be captured.
	Stdin bool `json:"stdin" yaml:"stdin" default:"false"`
	// Stdout signals that the standard output to the user should be captured.
	Stdout bool `json:"stdout" yaml:"stdout" default:"false"`
	// Stderr signals that the standard error to the user should be captured.
	Stderr bool `json:"stderr" yaml:"stderr" default:"false"`
	// Passwords signals that passwords during authentication should be captured.
	Passwords bool `json:"passwords" yaml:"passwords" default:"false"`
	// Forwarding signals that the contents of forward and reverse connection forwardings should be captured.
	Forwarding bool `json:"forwarding" yaml:"forwarding" default:"false"`
}

// Validate checks the configuration to enable global configuration check.
func (config *AuditLogConfig) Validate() error {
	if !config.Enable {
		return nil
	}
	if err := config.Format.Validate(); err != nil {
		return wrap(err, "format")
	}
	if err := config.Storage.Validate(); err != nil {
		return wrap(err, "storage")
	}
	switch config.Storage {
	case AuditLogStorageFile:
		return wrap(config.File.Validate(), "file")
	case AuditLogStorageS3:
		return wrap(config.S3.Validate(), "s3")
	}
	return nil
}

// AuditLogFileConfig is the configuration for the file storage.
type AuditLogFileConfig struct {
	Directory string `json:"directory" yaml:"directory" default:"/var/log/audit"`
}

func (c *AuditLogFileConfig) Validate() error {
	stat, err := os.Stat(c.Directory)
	if err != nil {
		return wrapWithMessage(err, "directory", "invalid audit log directory: %s", c.Directory)
	}
	if !stat.IsDir() {
		return newError("directory", "invalid audit log directory: %s (not a directory)", c.Directory)
	}
	return nil
}

// AuditLogS3Config S3 storage configuration
type AuditLogS3Config struct {
	Local           string             `json:"local" yaml:"local" default:"/var/lib/audit"`
	AccessKey       string             `json:"accessKey" yaml:"accessKey"`
	SecretKey       string             `json:"secretKey" yaml:"secretKey"`
	Bucket          string             `json:"bucket" yaml:"bucket"`
	Region          string             `json:"region" yaml:"region"`
	Endpoint        string             `json:"endpoint" yaml:"endpoint"`
	CACert          string             `json:"cacert" yaml:"cacert"`
	ACL             string             `json:"acl" yaml:"acl"`
	PathStyleAccess bool               `json:"pathStyleAccess" yaml:"pathStyleAccess"`
	UploadPartSize  uint               `json:"uploadPartSize" yaml:"uploadPartSize" default:"5242880"`
	ParallelUploads uint               `json:"parallelUploads" yaml:"parallelUploads" default:"20"`
	Metadata        AuditLogS3Metadata `json:"metadata" yaml:"metadata"`
}

// Validate validates the
func (config AuditLogS3Config) Validate() error {
	if config.Local == "" {
		return newError("local", "empty local storage directory provided")
	}
	stat, err := os.Stat(config.Local)
	if err != nil {
		return wrapWithMessage(err, "local", "invalid local directory: %s", config.Local)
	}
	if !stat.IsDir() {
		return newError("local", "invalid local directory: %s (not a directory)", config.Local)
	}
	if config.AccessKey == "" {
		return newError("accessKey", "no access key provided")
	}
	if config.SecretKey == "" {
		return newError("secretKey", "no secret key provided")
	}
	if config.Bucket == "" {
		return newError("bucket", "no bucket name provided")
	}
	if config.UploadPartSize < 5242880 {
		return newError("uploadPartSize", "upload part size too low %d (minimum 5 MB)", config.UploadPartSize)
	}
	if config.ParallelUploads < 1 {
		return newError("parallelUploads", "parallel uploads invalid: %d (must be positive)", config.ParallelUploads)
	}
	return nil
}

// AuditLogS3Metadata AuditLogS3Metadata configuration for the S3 storage
type AuditLogS3Metadata struct {
	IP       bool `json:"ip" yaml:"ip"`
	Username bool `json:"username" yaml:"username"`
}
