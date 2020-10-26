package s3

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"

	"github.com/containerssh/containerssh/audit"
	"github.com/containerssh/containerssh/config"
	"github.com/containerssh/containerssh/log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
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

	chunkSize := 5242880
	if cfg.UploadChunkSize > 5*1024*1024 {
		chunkSize = cfg.UploadChunkSize
	}

	sess := session.Must(session.NewSession(awsConfig))
	uploader := s3manager.NewUploader(sess)

	return &Storage{
		session:   sess,
		uploader:  uploader,
		logger:    logger,
		bucket:    cfg.Bucket,
		chunkSize: chunkSize,
	}, nil
}
