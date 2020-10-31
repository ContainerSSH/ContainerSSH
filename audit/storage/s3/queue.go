package s3

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/containerssh/containerssh/audit"
	"github.com/containerssh/containerssh/log"
	"io/ioutil"
	"os"
	"path"
	"sync"
)

var minPartSize = uint(5 * 1024 * 1024)
var maxPartSize = uint(5 * 1024 * 1024 * 1024)

type queueEntryMetadata struct {
	StartTime     int64  `json:"startTime" yaml:"startTime"`
	RemoteAddr    string `json:"remoteAddr" yaml:"remoteAddr"`
	Authenticated bool   `json:"authenticated" yaml:"authenticated"`
	Username      string `json:"username" yaml:"username"`
}

func (meta queueEntryMetadata) ToMap(showUsername bool, showIp bool) map[string]*string {
	var username *string
	if meta.Authenticated {
		username = &meta.Username
	}
	metadata := map[string]*string{
		"timestamp":     aws.String(fmt.Sprintf("%d", meta.StartTime)),
		"authenticated": aws.String(fmt.Sprintf("%t", meta.Authenticated)),
	}
	if showUsername {
		metadata["username"] = username
	}
	if showIp {
		metadata["ip"] = aws.String(meta.RemoteAddr)
	}
	return metadata
}

type queueEntry struct {
	logger        log.Logger
	name          string
	progress      int64
	finished      bool
	readHandle    *os.File
	writeHandle   *os.File
	partAvailable chan bool
	file          string
	metadata      queueEntryMetadata
}

// This method marks the the part as available if it has not been marked yet. This unfreezes the upload loop waiting in
// waitPartAvailable()
func (e *queueEntry) markPartAvailable() {
	select {
	case e.partAvailable <- true:
	default:
	}
}

// This method waits for the next part to be available.
func (e *queueEntry) waitPartAvailable() {
	<-e.partAvailable
}

func (e *queueEntry) remove() error {
	if e.readHandle != nil {
		if err := e.readHandle.Close(); err != nil {
			return fmt.Errorf("failed to close audit log file %s (%v)", e.file, err)
		} else {
			e.readHandle = nil
		}
	}
	if err := os.Remove(e.file); err != nil {
		return fmt.Errorf("failed to remove audit log file %s (%v)", e.name, err)
	}
	if err := os.Remove(fmt.Sprintf("%s.metadata.json", e.file)); err != nil {
		e.logger.WarningF("failed to remove audit log metadata file %s (%v)", e.name, err)
	}
	return nil
}

type uploadQueue struct {
	directory       string
	parallelUploads uint
	partSize        uint
	workerSem       chan int
	logger          log.Logger
	awsSession      *session.Session
	bucket          string
	acl             *string
	// queue map[string]*queueEntry
	queue            sync.Map
	metadataIp       bool
	metadataUsername bool
}

func newUploadQueue(
	directory string,
	partSize uint,
	parallelUploads uint,
	bucket string,
	acl string,
	metadataUsername bool,
	metadataIp bool,
	awsSession *session.Session,
	logger log.Logger,
) *uploadQueue {
	if partSize < minPartSize {
		partSize = minPartSize
	}
	if partSize > maxPartSize {
		partSize = maxPartSize
	}
	if parallelUploads < 1 {
		parallelUploads = 1
	}
	var realAcl *string = nil
	if acl != "" {
		realAcl = &acl
	}
	return &uploadQueue{
		directory:        directory,
		parallelUploads:  parallelUploads,
		partSize:         partSize,
		workerSem:        make(chan int, parallelUploads),
		logger:           logger,
		awsSession:       awsSession,
		bucket:           bucket,
		queue:            sync.Map{},
		acl:              realAcl,
		metadataIp:       metadataIp,
		metadataUsername: metadataUsername,
	}
}

func (q *uploadQueue) Open(name string) (audit.StorageWriter, error) {
	file := path.Join(q.directory, name)
	writeHandle, err := os.Create(file)
	if err != nil {
		return nil, err
	}
	readHandle, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	entry := &queueEntry{
		logger:        q.logger,
		name:          name,
		file:          file,
		progress:      0,
		finished:      false,
		readHandle:    readHandle,
		writeHandle:   writeHandle,
		partAvailable: make(chan bool, 1),
		metadata: queueEntryMetadata{
			StartTime:     0,
			RemoteAddr:    "",
			Authenticated: false,
			Username:      "",
		},
	}
	q.queue.Store(name, entry)
	err = q.upload(name)
	if err != nil {
		return nil, err
	}

	return newMonitoringWriter(
		writeHandle,
		q.partSize,
		func(startTime int64, remoteAddr string, username *string) {
			entry.metadata.StartTime = startTime
			entry.metadata.RemoteAddr = remoteAddr
			if username == nil {
				entry.metadata.Authenticated = false
				entry.metadata.Username = ""
			} else {
				entry.metadata.Authenticated = true
				entry.metadata.Username = *username
			}

			metadataFile := fmt.Sprintf("%s.metadata.json", file)
			metadataFileHandle, err := os.Create(metadataFile)
			if err != nil {
				q.logger.WarningF("failed to create local audit log %s metadata file (%v)", name, err)
			} else {
				defer func() {
					if err := metadataFileHandle.Close(); err != nil {
						q.logger.WarningF("failed to close audit log %s metadata file (%v)", name, err)
					}
				}()
				jsonData, err := json.Marshal(entry.metadata)
				if err != nil {
					q.logger.WarningF("failed to encode audit log %s metadata to JSON (%v)", name, err)
				} else {
					_, err := metadataFileHandle.Write(jsonData)
					if err != nil {
						q.logger.WarningF("failed to write audit log %s metadata file (%v)", name, err)
					}
				}
			}
		},
		func() {
			entry.markPartAvailable()
		},
		func() {
			err := q.finish(name)
			if err != nil {
				q.logger.WarningF("failed to finish audit log %s (%v)", name, err)
			}
		},
	), nil
}

func (q *uploadQueue) finish(name string) error {
	rawEntry, ok := q.queue.Load(name)
	if !ok {
		return fmt.Errorf("no such queue entry: %s", name)
	}
	entry := rawEntry.(*queueEntry)
	entry.finished = true
	entry.markPartAvailable()
	return nil
}

func (q *uploadQueue) recover(name string) error {
	file := path.Join(q.directory, name)

	readHandle, err := os.Open(file)
	if err != nil {
		return err
	}

	// Remove existing multipart upload
	if err = q.abortMultiPartUpload(name); err != nil {
		return fmt.Errorf("failed to recover upload for audit log %s (%v)", name, err)
	}
	metadata := &queueEntryMetadata{}
	metadataHandle, err := os.Open(fmt.Sprintf("%s.metadata.json", file))
	if err == nil {
		readBytes, err := ioutil.ReadAll(metadataHandle)
		if err != nil {
			q.logger.WarningF("failed to read audit log %s metadata file (%v)", name, err)
		} else {
			if err := json.Unmarshal(readBytes, metadata); err != nil {
				q.logger.WarningF("failed to unmarshal audit log %s JSON (%v)", name, err)
			}
		}
		if err := metadataHandle.Close(); err != nil {
			q.logger.WarningF("failed to close audit log %s metadata file (%v)", name, err)
		}
	} else {
		q.logger.WarningF("metadata file for recovered audit log %s failed to open (%v)", name, err)
	}

	// Create a new upload
	entry := &queueEntry{
		logger:        q.logger,
		name:          name,
		file:          file,
		progress:      0,
		finished:      true,
		readHandle:    readHandle,
		writeHandle:   nil,
		partAvailable: make(chan bool, 1),
		metadata:      *metadata,
	}
	q.queue.Store(name, entry)
	entry.markPartAvailable()
	return q.upload(name)
}
