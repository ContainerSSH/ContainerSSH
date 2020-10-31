package s3

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

func (q *uploadQueue) initializeMultiPartUpload(s3Connection *s3.S3, name string, metadata queueEntryMetadata) (*string, error) {
	q.logger.DebugF("initializing multipart upload for audit log %s...", name)
	multipartUpload, err := s3Connection.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
		Bucket:      aws.String(q.bucket),
		Key:         aws.String(name),
		ContentType: aws.String("application/octet-stream"),
		ACL:         q.acl,
		Metadata:    metadata.ToMap(),
	})
	if err != nil {
		q.logger.WarningF("failed to upload audit log file %s (%v)", name, err)
		return nil, err
	}
	return multipartUpload.UploadId, nil
}

func (q *uploadQueue) processMultiPartUploadPart(
	s3Connection *s3.S3,
	name string,
	uploadId string,
	partNumber int64,
	handle *os.File,
	startingByte int64,
	endingByte int64,
) (int64, string, error) {
	q.logger.DebugF("uploading part %d of audit log %s (part size %d bytes)...", partNumber, name, endingByte-startingByte)
	contentLength := endingByte - startingByte
	response, err := s3Connection.UploadPart(&s3.UploadPartInput{
		Body:          io.NewSectionReader(handle, startingByte, contentLength),
		Bucket:        aws.String(q.bucket),
		ContentLength: aws.Int64(contentLength),
		Key:           aws.String(name),
		PartNumber:    aws.Int64(partNumber),
		UploadId:      aws.String(uploadId),
	})
	etag := ""
	if err != nil {
		q.logger.WarningF("failed to upload part %d of audit log %s (%v)", partNumber, name, err)
		return 0, "", fmt.Errorf("failed to upload part %d of audit log file %s (%v)", partNumber, name, err)
	} else {
		etag = *response.ETag
	}
	q.logger.DebugF("completed upload of part %d of audit log %s", partNumber, name)
	return contentLength, etag, nil
}

func (q *uploadQueue) processSingleUpload(s3Connection *s3.S3, name string, handle *os.File, metadata queueEntryMetadata) (int64, error) {
	q.logger.DebugF("processing single upload for audit log %s...", name)
	stat, err := handle.Stat()
	if err != nil {
		return 0, fmt.Errorf("failed to upload audit log %s (%v)", name, err)
	}
	_, err = s3Connection.PutObject(&s3.PutObjectInput{
		Body:        handle,
		Bucket:      aws.String(q.bucket),
		Key:         aws.String(name),
		ContentType: aws.String("application/octet-stream"),
		ACL:         q.acl,
		Metadata:    metadata.ToMap(),
	})
	if err != nil {
		q.logger.DebugF("single upload failed for audit log %s (%v)", name, err)
	} else {
		q.logger.DebugF("single upload complete for audit log %s", name)
	}
	return stat.Size(), err
}

func (q *uploadQueue) finalizeUpload(s3Connection *s3.S3, name string, uploadId string, completedParts []*s3.CompletedPart) error {
	q.logger.DebugF("finalizing multipart upload for audit log %s...", name)
	_, err := s3Connection.CompleteMultipartUpload(&s3.CompleteMultipartUploadInput{
		Bucket: aws.String(q.bucket),
		Key:    aws.String(name),
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: completedParts,
		},
		UploadId: aws.String(uploadId),
	})
	if err != nil {
		q.logger.WarningF("finalizing multipart upload failed for audit log %s (%v)", name, err)
	} else {
		q.logger.DebugF("finalizing multipart upload complete for audit log %s", name)
	}
	return err
}

func (q *uploadQueue) upload(name string) error {
	_, ok := q.queue.Load(name)
	if !ok {
		return fmt.Errorf("no such queue entry: %s", name)
	}
	s3Connection := s3.New(q.awsSession)
	go func() {
		var uploadId *string = nil
		uploadedBytes := int64(0)
		errorHappened := false
		var completedParts []*s3.CompletedPart
		for {
			rawEntry, ok := q.queue.Load(name)
			if !ok {
				q.logger.WarningF("no such queue entry: %s", name)
				continue
			}
			entry := rawEntry.(*queueEntry)
			entry.waitPartAvailable()
			q.workerSem <- 42
			errorHappened = false

			remainingBytes := int64(-1)
			stat, err := entry.readHandle.Stat()
			if err == nil {
				remainingBytes = stat.Size() - uploadedBytes
			} else {
				q.logger.WarningF("failed to stat audit queue file %s before upload (%v)", name, err)
				errorHappened = true
			}

			if !errorHappened {
				if entry.finished && uploadedBytes == 0 {
					// If the entry is finished and nothing has been uploaded yet, upload it as a single file.
					partBytes, err := q.processSingleUpload(s3Connection, name, entry.readHandle, entry.metadata)
					if err != nil {
						q.logger.WarningF("failed to upload audit log %s (%v)", name, err)
						errorHappened = true
					} else {
						uploadedBytes = uploadedBytes + partBytes
					}
				} else if (entry.finished && remainingBytes > 0) || remainingBytes >= int64(q.partSize) {
					// If the entry is finished and there are bytes remaining, upload. Otherwise, we only upload if
					// more than the part size is available.
					if uploadId == nil {
						uploadId, err = q.initializeMultiPartUpload(s3Connection, name, entry.metadata)
						if err != nil {
							errorHappened = true
						}
					}
					if !errorHappened && uploadId != nil {
						partNumber := uploadedBytes / int64(q.partSize)
						startingByte := partNumber * int64(q.partSize)
						endingByte := (partNumber + 1) * int64(q.partSize)
						if stat.Size() < endingByte {
							endingByte = stat.Size()
						}

						if entry.finished && stat.Size()-endingByte < int64(q.partSize) {
							endingByte = stat.Size()
						}

						partBytes, etag, err := q.processMultiPartUploadPart(s3Connection, name, *uploadId, partNumber, entry.readHandle, startingByte, endingByte)
						if err == nil {
							uploadedBytes = uploadedBytes + partBytes
							completedParts = append(completedParts, &s3.CompletedPart{
								ETag:       aws.String(etag),
								PartNumber: aws.Int64(partNumber),
							})
						}
					}
				} else if entry.finished && remainingBytes == 0 {
					//If the entry is finished and no data is left to be uploaded, finalize the upload.
					if uploadId != nil {
						err := q.finalizeUpload(s3Connection, name, *uploadId, completedParts)
						if err != nil {
							errorHappened = true
						}
					}
					if !errorHappened {
						uploadId = nil
						if err := entry.remove(); err != nil {
							q.logger.WarningF("failed to remove queue entry (%v)", err)
						}
						q.queue.Delete(name)
						break
					}
				}
			}

			<-q.workerSem
			if errorHappened || (entry.finished && remainingBytes > 0) {
				// If an error happened, retry immediately.
				// Otherwise, if the entry is finished and there are remaining bytes, mark parts available for upload.
				entry.markPartAvailable()
			}
			if errorHappened {
				time.Sleep(10 * time.Second)
			}
		}
	}()
	return nil
}

func (q *uploadQueue) abortMultiPartUpload(name string) error {
	s3Connection := s3.New(q.awsSession)
	multiPartUpload, err := s3Connection.ListMultipartUploads(&s3.ListMultipartUploadsInput{
		Bucket: aws.String(q.bucket),
		Prefix: aws.String(name),
	})
	if err != nil {
		return fmt.Errorf("failed to list existing multipart upload for audit log %s (%v)", name, err)
	}
	for _, upload := range multiPartUpload.Uploads {
		if *upload.Key == name {
			q.logger.DebugF("aborting previous multipart upload ID %s for audit log %s...", *(upload.UploadId), name)
			_, err = s3Connection.AbortMultipartUpload(&s3.AbortMultipartUploadInput{
				Bucket:   aws.String(q.bucket),
				Key:      upload.Key,
				UploadId: upload.UploadId,
			})
			if err != nil {
				return fmt.Errorf("failed to abort  %s (%v)", name, err)
			}
		}
	}
	return nil
}
