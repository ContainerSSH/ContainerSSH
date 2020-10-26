package s3

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/containerssh/containerssh/log"
	"io"
)

type Storage struct {
	session   *session.Session
	uploader  *s3manager.Uploader
	logger    log.Logger
	bucket    string
	chunkSize int
}

func (s Storage) Open(name string) (io.WriteCloser, error) {
	reader, writer := io.Pipe()
	go func() {
		partNumber := 0
		buffer := new(bytes.Buffer)
		var uploadId *string = nil

		s3Connection := s3.New(s.session)
		initializeMultiPartUpload := func() error {
			multipartUpload, err := s3Connection.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
				Bucket:      aws.String(s.bucket),
				Key:         aws.String(name),
				ContentType: aws.String("application/octet-stream"),
			})
			if err != nil {
				s.logger.WarningF("failed to upload audit log file %s (%v)", name, err)
				return err
			}
			uploadId = multipartUpload.UploadId
			return nil
		}
		var finalizeUpload func() error
		uploadSingle := func() error {
			data := buffer.Bytes()
			buffer.Reset()
			contentLength := int64(len(data))
			_, err := s3Connection.PutObject(&s3.PutObjectInput{
				Body:          bytes.NewReader(data),
				Bucket:        aws.String(s.bucket),
				ContentLength: &contentLength,
				Key:           aws.String(name),
			})
			return err
		}
		uploadBuffer := func() error {
			if uploadId == nil {
				err := initializeMultiPartUpload()
				if err != nil {
					return err
				}
			}
			partBytes := buffer.Bytes()
			buffer.Reset()
			contentLength := int64(len(partBytes))
			_, err := s3Connection.UploadPart(&s3.UploadPartInput{
				Body:          bytes.NewReader(partBytes),
				Bucket:        aws.String(s.bucket),
				ContentLength: &contentLength,
				Key:           aws.String(name),
				PartNumber:    aws.Int64(int64(partNumber)),
				UploadId:      uploadId,
			})
			if err != nil {
				return fmt.Errorf("failed to upload part %d of audit log file %s (%v)", partNumber, name, err)
			}
			partNumber = partNumber + 1
			return nil
		}
		finalizeUpload = func() error {
			err := uploadBuffer()
			if err != nil {
				return err
			}
			_, err = s3Connection.CompleteMultipartUpload(&s3.CompleteMultipartUploadInput{
				Bucket:   aws.String(s.bucket),
				Key:      aws.String(name),
				UploadId: uploadId,
			})
			if err != nil {
				return err
			}
			return nil
		}
		for {
			data := make([]byte, 10*1024)
			readBytes, err := reader.Read(data)
			if err != nil {
				if err != io.EOF {
					s.logger.WarningF("failed to upload audit log file %s, attempting to finalize upload (%v)", name, err)
				}
				break
			}
			if readBytes == 0 {
				break
			}
			_, err = buffer.Write(data[0:readBytes])
			if buffer.Len() > s.chunkSize {
				err := uploadBuffer()
				if err != nil {
					s.logger.ErrorF("%v", err)
					err := finalizeUpload()
					if err != nil {
						s.logger.ErrorF("%v", err)
					}
					return
				}
			}
		}
		if buffer.Len() > 5*1024*1024 {
			err := uploadBuffer()
			if err != nil {
				s.logger.ErrorF("%v", err)
			}
		} else if buffer.Len() > 0 {
			err := uploadSingle()
			if err != nil {
				s.logger.ErrorF("%v", err)
			}
		}
		if uploadId != nil {
			err := finalizeUpload()
			if err != nil {
				s.logger.ErrorF("%v", err)
			}
		}
	}()

	return writer, nil
}
