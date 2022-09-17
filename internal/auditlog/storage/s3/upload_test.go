package s3_test

import (
	"context"
	"fmt"
	"io"
	goLog "log"
	"net"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	awsS3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"go.containerssh.io/libcontainerssh/config"
	auditLogStorage "go.containerssh.io/libcontainerssh/internal/auditlog/storage"
	"go.containerssh.io/libcontainerssh/internal/auditlog/storage/s3"
	"go.containerssh.io/libcontainerssh/log"
)

type minio struct {
	containerID string
	dir         string
	storage     auditLogStorage.ReadWriteStorage
}

func (m *minio) getClient() (*client.Client, error) {
	cli, err := client.NewClientWithOpts()
	if err != nil {
		return cli, err
	}
	cli.NegotiateAPIVersion(context.Background())
	return cli, nil
}

func (m *minio) Start(
	t *testing.T,
	accessKey string,
	secretKey string,
	region string,
	bucket string,
	endpoint string,
) (auditLogStorage.ReadWriteStorage, error) {
	if m.containerID == "" {
		if err := m.startMinio(t, accessKey, secretKey); err != nil {
			return nil, err
		}
	}

	var err error
	for i := 0; i < 6; i++ {
		if err = setupS3(accessKey, secretKey, endpoint, region, bucket); err != nil {
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}

	if err != nil {
		return nil, err
	}

	m.dir, err = os.MkdirTemp(os.TempDir(), "containerssh-s3-upload-test")
	if err != nil {
		assert.Fail(t, "failed to create temporary directory (%v)", err)
		return nil, err
	}

	err = m.setupStorage(t, accessKey, secretKey, region, bucket, endpoint)
	if err != nil {
		return nil, err
	}

	return m.storage, nil
}

func (m *minio) setupStorage(
	t *testing.T,
	accessKey string,
	secretKey string,
	region string,
	bucket string,
	endpoint string,
) error {
	logger := log.NewTestLogger(t)
	var err error
	m.storage, err = s3.NewStorage(
		config.AuditLogS3Config{
			Local:           m.dir,
			AccessKey:       accessKey,
			SecretKey:       secretKey,
			Bucket:          bucket,
			Region:          region,
			Endpoint:        endpoint,
			PathStyleAccess: true,
			UploadPartSize:  5 * 1024 * 1024,
			ParallelUploads: 20,
			Metadata:        config.AuditLogS3Metadata{},
		},
		logger,
	)
	if err != nil {
		assert.Fail(t, "failed to create storage (%v)", err)
		return err
	}
	return nil
}

var hostConfig = &container.HostConfig{
	AutoRemove: true,
	PortBindings: map[nat.Port][]nat.PortBinding{
		"9000/tcp": {
			{
				HostIP:   "127.0.0.1",
				HostPort: "9000",
			},
		},
	},
}

func (m *minio) startMinio(t *testing.T, accessKey string, secretKey string) error {
	ctx := context.Background()

	cli, err := m.getClient()
	if err != nil {
		assert.Fail(t, "failed to create Docker client (%v)", err)
		return err
	}

	reader, err := cli.ImagePull(ctx, "docker.io/minio/minio", types.ImagePullOptions{})
	if err != nil {
		assert.Fail(t, "failed to pull Minio image (%v)", err)
		return err
	}
	if _, err := io.Copy(os.Stdout, reader); err != nil {
		assert.Fail(t, "failed to stream logs from Minio image pull (%v)", err)
		return err
	}

	env := []string{
		fmt.Sprintf("MINIO_ACCESS_KEY=%s", accessKey),
		fmt.Sprintf("MINIO_SECRET_KEY=%s", secretKey),
	}

	resp, err := cli.ContainerCreate(
		ctx,
		&container.Config{
			Image: "minio/minio",
			Cmd:   []string{"server", "/testdata"},
			Env:   env,
		},
		hostConfig,
		nil,
		nil,
		"",
	)
	if err != nil {
		assert.Fail(t, "failed to create Minio container (%v)", err)
		return err
	}

	m.containerID = resp.ID

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		assert.Fail(t, "failed to start minio container (%v)", err)
		return err
	}

	if err := m.waitForMinio(t); err != nil {
		return err
	}

	return nil
}

func (m *minio) waitForMinio(t *testing.T) error {
	tries := 0
	for {
		if tries > 30 {
			m.Stop()
			assert.Fail(t, "minio failed to come up within 30 seconds")
			return fmt.Errorf("minio failed to come up")
		}
		sock, err := net.Dial("tcp", "127.0.0.1:9000")
		if err != nil {
			tries++
			time.Sleep(time.Second)
		} else {
			_ = sock.Close()

			break
		}
	}
	return nil
}

func (m *minio) Stop() {
	if m.containerID != "" {
		ctx := context.Background()

		cli, err := m.getClient()
		if err != nil {
			goLog.Printf("failed to create Docker client (%v)\n", err)
		}

		if err := cli.ContainerRemove(ctx, m.containerID, types.ContainerRemoveOptions{
			RemoveVolumes: false,
			RemoveLinks:   false,
			Force:         true,
		}); err != nil {
			goLog.Println("failed to stop Minio container", err)
		}

		ok, errChan := cli.ContainerWait(ctx, m.containerID, container.WaitConditionRemoved)
		select {
		case <-ok:
			m.containerID = ""
		case <-errChan:
		}
	}

	if m.dir != "" {
		m.dir = ""
		if err := os.RemoveAll(m.dir); err != nil {
			goLog.Printf("failed to remove temporary directory (%v)", err)
		}
	}
}

func setupS3(accessKey string, secretKey string, endpoint string, region string, bucket string) error {
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
	if _, err := s3Connection.CreateBucket(&awsS3.CreateBucketInput{
		Bucket: aws.String(bucket),
	}); err != nil {
		return err
	}
	return nil
}

func getS3Objects(t *testing.T, storage auditLogStorage.ReadWriteStorage) []auditLogStorage.Entry {
	var objects []auditLogStorage.Entry
	objectChan, errChan := storage.List()
	for {
		finished := false
		select {
		case object, ok := <-objectChan:
			if !ok {
				finished = true
				break
			}
			objects = append(objects, object)
		case err, ok := <-errChan:
			if !ok {
				finished = true
				break
			}
			assert.Fail(t, "error while fetching objects from object storage", err)
		}
		if finished {
			break
		}
	}
	return objects
}

func waitForS3Objects(t *testing.T, storage auditLogStorage.ReadWriteStorage, count int) []auditLogStorage.Entry {
	tries := 0
	var objects []auditLogStorage.Entry
	for {
		if tries > 10 {
			break
		}
		objects = getS3Objects(t, storage)
		if len(objects) > count-1 {
			break
		} else {
			tries++
			time.Sleep(10 * time.Second)
		}
	}
	return objects
}

func TestSmallUpload(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping test in short mode")
	}
	accessKey := "asdfasdfasdf"
	secretKey := "asdfasdfasdf"

	region := "us-east-1"
	bucket := "auditlog"
	endpoint := "http://127.0.0.1:9000"

	m := &minio{}

	storage, err := m.Start(t, accessKey, secretKey, region, bucket, endpoint)
	defer m.Stop()
	if err != nil {
		return
	}

	writer, err := storage.OpenWriter("test")
	if err != nil {
		assert.Fail(t, "failed to open storage writer (%v)", err)
		return
	}
	var data = []byte("Hello world!")
	if _, err := writer.Write(data); err != nil {
		assert.Fail(t, "failed to write to storage writer (%v)", err)
		return
	}
	if err := writer.Close(); err != nil {
		assert.Fail(t, "failed to close storage writer (%v)", err)
		return
	}

	storage.Shutdown(context.Background())

	objects := waitForS3Objects(t, storage, 1)
	assert.Equal(t, 1, len(objects))

	r, err := storage.OpenReader(objects[0].Name)
	if err != nil {
		assert.Fail(t, "failed to open reader for recently stored object", err)
		return
	}
	d, err := io.ReadAll(r)
	if err != nil {
		assert.Fail(t, "failed to open read from S3", err)
		return
	}

	assert.Equal(t, data, d)
}

func TestLargeUpload(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping test in short mode")
	}
	accessKey := "asdfasdfasdf"
	secretKey := "asdfasdfasdf"

	region := "us-east-1"
	bucket := "auditlog"
	endpoint := "http://127.0.0.1:9000"

	m := &minio{}

	storage, err := m.Start(t, accessKey, secretKey, region, bucket, endpoint)
	defer m.Stop()
	if err != nil {
		return
	}

	writer, err := storage.OpenWriter("test")
	if err != nil {
		assert.Fail(t, "failed to open storage writer (%v)", err)
		return
	}
	size := 25 * 1000 * 1000
	var data = []byte("0123456789")
	for i := 0; i < size/len(data); i++ {
		if _, err := writer.Write(data); err != nil {
			assert.Fail(t, "failed to write to storage writer (%v)", err)
			return
		}
	}
	if err := writer.Close(); err != nil {
		assert.Fail(t, "failed to close storage writer (%v)", err)
		return
	}

	storage.Shutdown(context.Background())

	objects := waitForS3Objects(t, storage, 1)
	assert.Equal(t, 1, len(objects))

	r, err := storage.OpenReader(objects[0].Name)
	if err != nil {
		assert.Fail(t, "failed to open reader for recently stored object", err)
		return
	}
	d, err := io.ReadAll(r)
	if err != nil {
		assert.Fail(t, "failed to open read from S3", err)
		return
	}

	assert.Equal(t, size, len(d))
}
