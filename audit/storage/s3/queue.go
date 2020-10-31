package s3

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/containerssh/containerssh/log"
	"io"
	"os"
	"path"
	"sync"
)

var minPartSize = uint(5 * 1024 * 1024)
var maxPartSize = uint(5 * 1024 * 1024 * 1024)

type queueEntry struct {
	name          string
	progress      int64
	finished      bool
	readHandle    *os.File
	writeHandle   *os.File
	partAvailable chan bool
	file          string
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
	queue sync.Map
}

func newUploadQueue(
	directory string,
	partSize uint,
	parallelUploads uint,
	bucket string,
	acl string,
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
		directory:       directory,
		parallelUploads: parallelUploads,
		partSize:        partSize,
		workerSem:       make(chan int, parallelUploads),
		logger:          logger,
		awsSession:      awsSession,
		bucket:          bucket,
		queue:           sync.Map{},
		acl:             realAcl,
	}
}

func (q *uploadQueue) Open(name string) (io.WriteCloser, error) {
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
		name:          name,
		file:          file,
		progress:      0,
		finished:      false,
		readHandle:    readHandle,
		writeHandle:   writeHandle,
		partAvailable: make(chan bool, 1),
	}
	q.queue.Store(name, entry)
	err = q.upload(name)
	if err != nil {
		return nil, err
	}

	return newMonitoringWriter(
		writeHandle,
		q.partSize,
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

	// Create a new upload
	entry := &queueEntry{
		name:          name,
		file:          file,
		progress:      0,
		finished:      true,
		readHandle:    readHandle,
		writeHandle:   nil,
		partAvailable: make(chan bool, 1),
	}
	q.queue.Store(name, entry)
	entry.markPartAvailable()
	return q.upload(name)
}
