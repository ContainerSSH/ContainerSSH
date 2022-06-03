package s3

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	awsS3 "github.com/aws/aws-sdk-go/service/s3"

    "go.containerssh.io/libcontainerssh/internal/auditlog/storage"
)

func (q *uploadQueue) List() (<-chan storage.Entry, <-chan error) {
	s3Connection := awsS3.New(q.awsSession)

	result := make(chan storage.Entry)
	errChannel := make(chan error)

	go func() {
		var continuationToken *string = nil
		for {
			listObjectsResult, err := s3Connection.ListObjectsV2(
				&awsS3.ListObjectsV2Input{
					Bucket:            aws.String(q.bucket),
					ContinuationToken: continuationToken,
				},
			)
			if err != nil {
				errChannel <- err
				close(result)
				close(errChannel)
				return
			}
			for _, object := range listObjectsResult.Contents {
				name := object.Key

				headObjectResult, err := s3Connection.HeadObject(
					&awsS3.HeadObjectInput{
						Bucket: aws.String(q.bucket),
						Key:    name,
					},
				)
				if err != nil {
					errChannel <- fmt.Errorf("failed to fetch metadata for audit log %s (%w)", *name, err)
					continue
				}

				meta := map[string]string{}
				for k, v := range headObjectResult.Metadata {
					if v != nil {
						meta[k] = *v
					}
				}

				result <- storage.Entry{
					Name:     *name,
					Metadata: meta,
				}
			}
			continuationToken = listObjectsResult.NextContinuationToken
			if continuationToken == nil {
				break
			}
		}
		close(result)
		close(errChannel)
	}()

	return result, errChannel
}
