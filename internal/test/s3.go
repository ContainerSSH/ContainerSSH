package test

import (
	"fmt"
	"net"
	"testing"
	"time"
)

// s3Lock is used to start a maximum of 5 Minio instances for testing.
var s3Lock = make(chan struct{}, 5)

// S3 starts up an S3-compatible object storage using Docker for testing, and returns an object that
// can be queried for connection parameters. When the test finishes it automatically tears down the object storage.
//
// The connection parameters can be fetched from the returned helper object.
func S3(t *testing.T) S3Helper {
	t.Helper()

	s3Lock <- struct{}{}
	t.Cleanup(func() {
		<-s3Lock
	})

	accessKey := "test"
	secretKey := "testtest"
	env := []string{
		fmt.Sprintf("MINIO_ROOT_USER=%s", accessKey),
		fmt.Sprintf("MINIO_ROOT_PASSWORD=%s", secretKey),
	}
	t.Log("Starting Minio in a container...")
	m := &minio{
		cnt: containerFromPull(
			t,
			"docker.io/minio/minio",
			[]string{"server", "/testdata"},
			env,
			map[string]string{
				"9000/tcp": "",
			},
		),
		accessKey: accessKey,
		secretKey: secretKey,
		t:         t,
	}
	m.wait()

	m.t.Logf("Minio is now available at 127.0.0.1:%d.", m.cnt.port("9000/tcp"))

	return m
}

// S3Helper gives access to an S3-compatible object storage.
type S3Helper interface {
	// URL returns the endpoint for the S3 connection.
	URL() string
	// AccessKey returns the access key ID that can be used to access the S3 service.
	AccessKey() string
	// SecretKey returns the secret access key that can be used to access the S3 service.
	SecretKey() string
	// Region returns the S3 region string to use.
	Region() string
	// PathStyle returns true if path-style access should be used.
	PathStyle() bool
}

type minio struct {
	cnt       container
	accessKey string
	secretKey string
	t         *testing.T
}

func (m *minio) PathStyle() bool {
	return true
}

func (m *minio) Region() string {
	return "us-east-1"
}

func (m *minio) URL() string {
	return fmt.Sprintf("http://127.0.0.1:%d/", m.cnt.port("9000/tcp"))
}

func (m *minio) AccessKey() string {
	return m.accessKey
}

func (m *minio) SecretKey() string {
	return m.secretKey
}

func (m *minio) wait() {
	m.t.Log("Waiting for Minio to come up...")
	tries := 0
	sleepTime := 5
	for {
		if tries > 30 {
			m.t.Fatalf("Minio failed to come up in %d seconds.", sleepTime*30)
		}
		sock, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", m.cnt.port("9000/tcp")))
		time.Sleep(time.Duration(sleepTime) * time.Second)
		if err != nil {
			tries++
		} else {
			_ = sock.Close()

			return
		}
	}
}
