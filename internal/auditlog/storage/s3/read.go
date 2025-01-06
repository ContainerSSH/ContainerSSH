package s3

import (
	"io"

	"github.com/aws/aws-sdk-go/aws"
	awsS3 "github.com/aws/aws-sdk-go/service/s3"
)

func (q *uploadQueue) OpenReader(name string) (io.ReadCloser, error) {
	s3Connection := awsS3.New(q.awsSession)

	getObjectOutput, err := s3Connection.GetObject(&awsS3.GetObjectInput{
		Bucket: aws.String(q.bucket),
		Key:    aws.String(name),
	})
	if err != nil {
		return nil, err
	}

	return getObjectOutput.Body, nil
}
