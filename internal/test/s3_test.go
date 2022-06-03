package test_test

import (
	"io"
	"io/ioutil"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
    "go.containerssh.io/libcontainerssh/internal/test"
)

func TestS3(t *testing.T) {
	s3Service := test.S3(t)

	awsConfig := &aws.Config{
		Credentials: credentials.NewCredentials(
			&credentials.StaticProvider{
				Value: credentials.Value{
					AccessKeyID:     s3Service.AccessKey(),
					SecretAccessKey: s3Service.SecretKey(),
				},
			},
		),
		Endpoint:         aws.String(s3Service.URL()),
		Region:           aws.String(s3Service.Region()),
		S3ForcePathStyle: aws.Bool(s3Service.PathStyle()),
	}
	sess, err := session.NewSession(awsConfig)
	if err != nil {
		t.Fatalf("failed to establish S3 session (%v)", err)
	}
	s3Connection := s3.New(sess)

	s3CreateTestBucket(t, s3Connection)
	s3UploadTestObject(t, s3Connection)
	s3ReadAndCheckTestObject(t, s3Connection)
}

func s3CreateTestBucket(t *testing.T, s3Connection *s3.S3) {
	t.Log("Creating test bucket...")
	if _, err := s3Connection.CreateBucket(
		&s3.CreateBucketInput{
			Bucket: aws.String("test"),
		},
	); err != nil {
		t.Fatalf("failed to create bucket (%v)", err)
	}
}

func s3UploadTestObject(t *testing.T, s3Connection *s3.S3) {
	t.Log("Creating test object...")
	if _, err := s3Connection.PutObject(
		&s3.PutObjectInput{
			Key:    aws.String("test.txt"),
			Body:   &byteReader{data: []byte("Hello world!")},
			Bucket: aws.String("test"),
		},
	); err != nil {
		t.Fatalf("failed to put object (%v)", err)
	}
}

func s3ReadAndCheckTestObject(t *testing.T, s3Connection *s3.S3) {
	t.Log("Getting test object...")
	getObjectResponse, err := s3Connection.GetObject(
		&s3.GetObjectInput{
			Bucket: aws.String("test"),
			Key:    aws.String("test.txt"),
		},
	)
	if err != nil {
		t.Fatalf("failed to get object (%v)", err)
	}
	defer func() {
		_ = getObjectResponse.Body.Close()
	}()
	data, err := ioutil.ReadAll(getObjectResponse.Body)
	if err != nil {
		t.Fatalf("failed to read get object response body (%v)", err)
	}
	if string(data) != "Hello world!" {
		t.Fatalf("incorrect testdata in S3 bucket: %s", data)
	}
}

type byteReader struct {
	cursor  int64
	data    []byte
	current []byte
}

func (b *byteReader) Read(p []byte) (n int, err error) {
	n = copy(p, b.current)
	if n == 0 {
		return n, io.EOF
	}
	b.current = b.current[n:]
	b.cursor += int64(n)
	return
}

func (b *byteReader) Seek(offset int64, whence int) (int64, error) {
	var newCursor int64
	switch whence {
	case io.SeekStart:
		newCursor = offset
	case io.SeekEnd:
		newCursor = int64(len(b.data)) + offset
	case io.SeekCurrent:
		newCursor = b.cursor + offset
	}
	if newCursor < 0 {
		newCursor = 0
	}
	if newCursor > int64(len(b.data)) {
		newCursor = int64(len(b.data))
	}
	b.cursor = newCursor
	b.current = b.data[newCursor:]
	return b.cursor, nil
}
