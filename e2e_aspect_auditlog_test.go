package containerssh_test

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	awsS3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/containerssh/libcontainerssh/config"
	"github.com/containerssh/libcontainerssh/log"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

func NewAuditLogTestingAspect() TestingAspect {
	return &auditLogAspect{
		lock: &sync.Mutex{},
	}
}

type auditLogAspect struct {
	lock *sync.Mutex
}

func (a *auditLogAspect) String() string {
	return "Auditlog AuditLogStorage"
}

func (a *auditLogAspect) Factors() []TestingFactor {
	cli, err := client.NewClientWithOpts()
	if err != nil {
		panic(err)
	}

	cli.NegotiateAPIVersion(context.Background())

	return []TestingFactor{
		&auditLogFactor{
			aspect:  a,
			storage: config.AuditLogStorageNone,
			lock:    a.lock,
		},
		&auditLogFactor{
			aspect:  a,
			storage: config.AuditLogStorageFile,
			lock:    a.lock,
		},
		&auditLogFactor{
			aspect:       a,
			storage:      config.AuditLogStorageS3,
			dockerClient: cli,
			lock:         a.lock,
		},
	}
}

type auditLogFactor struct {
	aspect       *auditLogAspect
	storage      config.AuditLogStorage
	dockerClient *client.Client
	containerID  string
	lock         *sync.Mutex
}

func (a *auditLogFactor) Aspect() TestingAspect {
	return a.aspect
}

func (a *auditLogFactor) String() string {
	return string(a.storage)
}

func (a *auditLogFactor) ModifyConfiguration(cfg *config.AppConfig) error {
	switch a.storage {
	case config.AuditLogStorageNone:
		cfg.Audit.Enable = false
	case config.AuditLogStorageFile:
		cfg.Audit.Enable = true
		tmpDir, err := ioutil.TempDir(os.TempDir(), "containerssh-audit-*")
		if err != nil {
			return err
		}

		cfg.Audit.File.Directory = tmpDir
	case config.AuditLogStorageS3:
		cfg.Audit.Enable = true
		cfg.Audit.S3.AccessKey = "auditlog"
		cfg.Audit.S3.SecretKey = "auditlog"
		cfg.Audit.S3.PathStyleAccess = true
		cfg.Audit.S3.Bucket = "auditlog"
		cfg.Audit.S3.Region = "us-east-1"
		cfg.Audit.S3.Endpoint = "http://127.0.0.1:9000"
		tmpDir, err := ioutil.TempDir(os.TempDir(), "containerssh-audit-*")
		if err != nil {
			return err
		}
		cfg.Audit.S3.Local = tmpDir
	}

	cfg.Audit.Storage = a.storage
	cfg.Audit.Format = config.AuditLogFormatBinary
	return nil
}

func (a *auditLogFactor) StartBackingServices(
	cfg config.AppConfig,
	_ log.Logger,
) error {
	if cfg.Audit.Storage != config.AuditLogStorageS3 {
		return nil
	}
	a.lock.Lock()
	reader, err := a.dockerClient.ImagePull(
		context.Background(),
		"docker.io/minio/minio",
		types.ImagePullOptions{},
	)
	if err != nil {
		return err
	}
	if _, err := io.Copy(os.Stdout, reader); err != nil {
		return err
	}
	resp, err := a.createMinio(cfg)
	if err != nil {
		return err
	}
	a.containerID = resp.ID

	if err := a.dockerClient.ContainerStart(
		context.Background(),
		a.containerID,
		types.ContainerStartOptions{},
	); err != nil {
		_ = a.dockerClient.ContainerRemove(
			context.Background(), a.containerID, types.ContainerRemoveOptions{
				Force: true,
			},
		)
		return err
	}

	if err := a.waitForMinio(); err != nil {
		return err
	}

	if err := a.setupS3(
		cfg.Audit.S3.AccessKey,
		cfg.Audit.S3.SecretKey,
		cfg.Audit.S3.Endpoint,
		cfg.Audit.S3.Region,
		cfg.Audit.S3.Bucket,
	); err != nil {
		return err
	}

	time.Sleep(10 * time.Second)

	return nil
}

func (a *auditLogFactor) setupS3(accessKey string, secretKey string, endpoint string, region string, bucket string) error {
	awsConfig := &aws.Config{
		CredentialsChainVerboseErrors: nil,
		Credentials: credentials.NewCredentials(&credentials.StaticProvider{
			Value: credentials.Value{
				AccessKeyID:     accessKey,
				SecretAccessKey: secretKey,
			},
		}),
		Endpoint:         aws.String(endpoint),
		Region:           aws.String(region),
		S3ForcePathStyle: aws.Bool(true),
	}
	sess := session.Must(session.NewSession(awsConfig))
	s3Connection := awsS3.New(sess)
	tries := 0
	var lastError error
	for {
		if tries > 30 {
			return lastError
		}
		if _, lastError = s3Connection.CreateBucket(
			&awsS3.CreateBucketInput{
				Bucket: aws.String(bucket),
			},
		); lastError != nil {
			tries++
			time.Sleep(time.Second)
		} else {
			return nil
		}
	}
}

func (a *auditLogFactor) createMinio(cfg config.AppConfig) (container.ContainerCreateCreatedBody, error) {
	env := []string{
		fmt.Sprintf("MINIO_ACCESS_KEY=%s", cfg.Audit.S3.AccessKey),
		fmt.Sprintf("MINIO_SECRET_KEY=%s", cfg.Audit.S3.AccessKey),
	}

	return a.dockerClient.ContainerCreate(
		context.Background(),
		&container.Config{
			Image: "minio/minio",
			Cmd:   []string{"server", "/testdata"},
			Env:   env,
		},
		&container.HostConfig{
			PortBindings: map[nat.Port][]nat.PortBinding{
				"9000/tcp": {
					{
						HostIP:   "127.0.0.1",
						HostPort: "9000",
					},
				},
			},
			AutoRemove: true,
		},
		nil,
		nil,
		"",
	)
}

func (a *auditLogFactor) waitForMinio() error {
	tries := 0
	for {
		if tries > 30 {
			timeout := 30 * time.Second
			_ = a.dockerClient.ContainerStop(context.Background(), a.containerID, &timeout)
			return fmt.Errorf("minio failed to come up within 30 seconds")
		}
		inspectResult, err := a.dockerClient.ContainerInspect(context.Background(), a.containerID)
		if err != nil {
			tries++
			time.Sleep(time.Second)
			continue
		}
		if inspectResult.State != nil {
			if !inspectResult.State.Running {
				tries++
				time.Sleep(time.Second)
				continue
			}
		}

		sock, err := net.Dial("tcp", "127.0.0.1:9000")
		if err != nil {
			tries++
			time.Sleep(time.Second)
			continue
		} else {
			_ = sock.Close()

			break
		}
	}
	return nil
}

func (a *auditLogFactor) StopBackingServices(
	cfg config.AppConfig,
	_ log.Logger,
) error {
	if a.storage == config.AuditLogStorageFile {
		return os.RemoveAll(cfg.Audit.File.Directory)
	} else if a.storage == config.AuditLogStorageNone {
		return nil
	} else {
		defer a.lock.Unlock()
		timeout := 30 * time.Second
		if err := a.dockerClient.ContainerStop(context.Background(), a.containerID, &timeout); err != nil {
			return err
		}
		return os.RemoveAll(cfg.Audit.S3.Local)
	}
}

func (a *auditLogFactor) GetSteps(
	_ config.AppConfig,
	_ log.Logger,
) []Step {
	return []Step{}
}
