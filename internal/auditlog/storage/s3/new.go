package s3

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
    "go.containerssh.io/libcontainerssh/config"
    "go.containerssh.io/libcontainerssh/log"

    "go.containerssh.io/libcontainerssh/internal/auditlog/storage"
)

// NewStorage Creates a storage driver for an S3-compatible object storage.
func NewStorage(cfg config.AuditLogS3Config, logger log.Logger) (storage.ReadWriteStorage, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	httpClient, err := getHTTPClient(cfg)
	if err != nil {
		return nil, err
	}

	awsConfig := getAWSConfig(cfg, logger, httpClient)

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, err
	}

	queue := newUploadQueue(
		cfg.Local,
		cfg.UploadPartSize,
		cfg.ParallelUploads,
		cfg.Bucket,
		cfg.ACL,
		cfg.Metadata.Username,
		cfg.Metadata.IP,
		sess,
		logger,
	)

	if _, err := os.Stat(cfg.Local); err != nil {
		return nil, fmt.Errorf("invalid local audit directory %s (%w)", cfg.Local, err)
	}

	if err := filepath.Walk(cfg.Local, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && info.Size() > 0 && !strings.Contains(info.Name(), ".") {
			if err := queue.recover(info.Name()); err != nil {
				return fmt.Errorf("failed to enqueue old audit log file %s (%w)", info.Name(), err)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return queue, nil
}

func getAWSConfig(
	cfg config.AuditLogS3Config, logger log.Logger, httpClient *http.Client,
) *aws.Config {
	var endpoint *string
	if cfg.Endpoint != "" {
		endpoint = &cfg.Endpoint
	}

	awsConfig := &aws.Config{
		Credentials: credentials.NewCredentials(&credentials.StaticProvider{
			Value: credentials.Value{
				AccessKeyID:     cfg.AccessKey,
				SecretAccessKey: cfg.SecretKey,

				SessionToken: "",
				ProviderName: "",
			},
		}),
		Endpoint:         endpoint,
		Region:           &cfg.Region,
		HTTPClient:       httpClient,
		Logger:           logger,
		S3ForcePathStyle: aws.Bool(cfg.PathStyleAccess),
	}

	return awsConfig
}

func getHTTPClient(cfg config.AuditLogS3Config) (*http.Client, error) {
	httpClient := http.DefaultClient
	if cfg.CACert != "" {
		rootCAs, _ := x509.SystemCertPool()
		if rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}
		if ok := rootCAs.AppendCertsFromPEM([]byte(cfg.CACert)); !ok {
			return nil, fmt.Errorf("failed to add certificate from config file")
		}
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS13,
			RootCAs:    rootCAs,
		}
		httpTransport := &http.Transport{
			TLSClientConfig: tlsConfig,
		}
		httpClient = &http.Client{
			Transport:     httpTransport,
			CheckRedirect: nil,
			Jar:           nil,
			Timeout:       0,
		}
	}
	return httpClient, nil
}
