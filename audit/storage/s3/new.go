package s3

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/containerssh/containerssh/audit"
	"github.com/containerssh/containerssh/config"
	"github.com/containerssh/containerssh/log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

func NewStorage(cfg config.AuditS3Config, logger log.Logger) (audit.Storage, error) {
	httpClient := http.DefaultClient
	if cfg.CaCert != "" {
		rootCAs, _ := x509.SystemCertPool()
		if rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}
		if ok := rootCAs.AppendCertsFromPEM([]byte(cfg.CaCert)); !ok {
			return nil, fmt.Errorf("failed to add certificate from config file")
		}
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
			RootCAs:            rootCAs,
		}
		httpTransport := &http.Transport{TLSClientConfig: tlsConfig}
		httpClient = &http.Client{Transport: httpTransport}
	}

	var endpoint *string
	if cfg.Endpoint != "" {
		endpoint = &cfg.Endpoint
	}

	if cfg.Bucket == "" {
		return nil, fmt.Errorf("no bucket name specified")
	}
	if cfg.Region == "" {
		return nil, fmt.Errorf("no region name specified")
	}

	awsConfig := &aws.Config{
		Credentials: credentials.NewCredentials(&credentials.StaticProvider{
			Value: credentials.Value{
				AccessKeyID:     cfg.AccessKey,
				SecretAccessKey: cfg.SecretKey,
			},
		}),
		Endpoint:   endpoint,
		Region:     &cfg.Region,
		Logger:     logger,
		HTTPClient: httpClient,
	}

	partSize := uint(5242880)
	if cfg.UploadPartSize > 5242880 {
		partSize = cfg.UploadPartSize
	}
	parallelUploads := uint(20)
	if cfg.ParallelUploads > 1 {
		parallelUploads = cfg.ParallelUploads
	}

	sess := session.Must(session.NewSession(awsConfig))

	queue := newUploadQueue(
		cfg.Local,
		partSize,
		parallelUploads,
		cfg.Bucket,
		cfg.ACL,
		sess,
		logger,
	)

	if _, err := os.Stat(cfg.Local); err != nil {
		return nil, fmt.Errorf("invalid local audit directory %s (%v)", cfg.Local, err)
	}

	if err := filepath.Walk(cfg.Local, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && info.Size() > 0 && !strings.Contains(info.Name(), ".") {
			if err := queue.recover(info.Name()); err != nil {
				return fmt.Errorf("failed to enqueue old audit log file %s (%v)", info.Name(), err)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return queue, nil
}
