package s3

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
    "go.containerssh.io/libcontainerssh/internal/auditlog/storage"
    "go.containerssh.io/libcontainerssh/log"
    "go.containerssh.io/libcontainerssh/message"
)

var minPartSize = uint(5 * 1024 * 1024)
var maxPartSize = uint(5 * 1024 * 1024 * 1024)

type queueEntryMetadata struct {
	StartTime     int64  `json:"startTime" yaml:"startTime"`
	RemoteAddr    string `json:"remoteAddr" yaml:"remoteAddr"`
	Authenticated bool   `json:"authenticated" yaml:"authenticated"`
	Username      string `json:"username" yaml:"username"`
	Country       string `json:"country" yaml:"country"`
}

func (meta queueEntryMetadata) ToMap(showUsername bool, showIP bool) map[string]*string {
	var username *string
	if meta.Authenticated {
		username = &meta.Username
	}
	metadata := map[string]*string{
		"timestamp":     aws.String(fmt.Sprintf("%d", meta.StartTime)),
		"authenticated": aws.String(fmt.Sprintf("%t", meta.Authenticated)),
		"country":       aws.String(meta.Country),
	}
	if showUsername {
		metadata["username"] = username
	}
	if showIP {
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
	lock          *sync.Mutex
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
			return message.Wrap(
				err,
				message.EAuditLogCloseAuditLogFileFailed,
				"failed to close audit log file %s (%w)",
				e.file,
			)
		}
		e.readHandle = nil
	}
	if err := os.Remove(e.file); err != nil {
		return message.Wrap(err, message.EAuditLogRemoveAuditLogFailed, "failed to remove audit log file %s", e.name)
	}
	metadataFile := fmt.Sprintf("%s.metadata.json", e.file)
	if _, err := os.Stat(metadataFile); err == nil {
		if err := os.Remove(metadataFile); err != nil {
			e.logger.Warning(message.Wrap(err, "failed to remove audit log metadata file %s", e.name))
		}
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
	metadataIP       bool
	metadataUsername bool
	wg               *sync.WaitGroup
	ctx              context.Context
	cancelFunc       context.CancelFunc
	shutdownContext  context.Context
	lock             *sync.Mutex
}

func (q *uploadQueue) Shutdown(shutdownContext context.Context) {
	q.lock.Lock()
	q.shutdownContext = shutdownContext
	q.cancelFunc()
	q.lock.Unlock()
	q.wg.Wait()
}

func newUploadQueue(
	directory string,
	partSize uint,
	parallelUploads uint,
	bucket string,
	acl string,
	metadataUsername bool,
	metadataIP bool,
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
	var realACL *string = nil
	if acl != "" {
		realACL = &acl
	}
	ctx, cancelFunc := context.WithCancel(context.Background())
	return &uploadQueue{
		lock:             &sync.Mutex{},
		directory:        directory,
		parallelUploads:  parallelUploads,
		partSize:         partSize,
		workerSem:        make(chan int, parallelUploads),
		logger:           logger,
		awsSession:       awsSession,
		bucket:           bucket,
		queue:            sync.Map{},
		acl:              realACL,
		metadataIP:       metadataIP,
		metadataUsername: metadataUsername,
		wg:               &sync.WaitGroup{},
		ctx:              ctx,
		cancelFunc:       cancelFunc,
	}
}

func (q *uploadQueue) OpenWriter(name string) (storage.Writer, error) {
	file := path.Join(q.directory, name)
	writeHandle, err := os.Create(file)
	if err != nil {
		return nil, err
	}
	// We are deliberately opening a file here.
	readHandle, err := os.Open(file) //nolint:gosec
	if err != nil {
		return nil, err
	}
	entry := &queueEntry{
		lock:          &sync.Mutex{},
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

	return q.getMonitoringWriter(name, writeHandle, entry, file), nil
}

func (q *uploadQueue) getMonitoringWriter(
	name string,
	writeHandle *os.File,
	entry *queueEntry,
	file string,
) storage.Writer {
	return newMonitoringWriter(
		writeHandle,
		q.partSize,
		func(startTime int64, remoteAddr string, country string, username *string) {
			entry.metadata.StartTime = startTime
			entry.metadata.RemoteAddr = remoteAddr
			entry.metadata.Country = country
			if username == nil {
				entry.metadata.Authenticated = false
				entry.metadata.Username = ""
			} else {
				entry.metadata.Authenticated = true
				entry.metadata.Username = *username
			}

			metadataFile := fmt.Sprintf("%s.metadata.json", file)
			q.writeMetadataFile(metadataFile, name, entry)
		},
		func() {
			entry.markPartAvailable()
		},
		func() {
			err := q.finish(name)
			if err != nil {
				q.logger.Warning(err)
			}
		},
	)
}

func (q *uploadQueue) writeMetadataFile(metadataFile string, name string, entry *queueEntry) {
	metadataFileHandle, err := os.Create(metadataFile)
	if err != nil {
		q.logger.Warning(
			message.Wrap(
				err,
				message.EAuditLogFailedCreatingMetadataFile,
				"failed to create local audit log %s metadata file",
				name,
			),
		)
	} else {
		defer func() {
			if err := metadataFileHandle.Close(); err != nil {
				q.logger.Warning(
					message.Wrap(
						err,
						message.EAuditLogCannotCloseMetadataFileHandle,
						"failed to close audit log %s metadata file",
						name,
					),
				)
			}
		}()
		jsonData, err := json.Marshal(entry.metadata)
		if err != nil {
			q.logger.Warning(
				message.Wrap(
					err,
					message.EAuditLogFailedMetadataJSONEncoding,
					"failed to encode audit log %s metadata to JSON",
					name,
				),
			)
		} else {
			_, err := metadataFileHandle.Write(jsonData)
			if err != nil {
				q.logger.Warning(
					message.Wrap(
						err,
						message.EAuditLogFailedWritingMetadataFile,
						"failed to write audit log %s metadata file",
						name,
					),
				)
			}
		}
	}
}

func (q *uploadQueue) finish(name string) error {
	rawEntry, ok := q.queue.Load(name)
	if !ok {
		return message.NewMessage(message.EAuditLogNoSuchQueueEntry, "no such queue entry: %s", name)
	}
	entry := rawEntry.(*queueEntry)
	entry.lock.Lock()
	entry.finished = true
	entry.markPartAvailable()
	entry.lock.Unlock()
	return nil
}

func (q *uploadQueue) recover(name string) error {
	q.logger.Debug(
		message.NewMessage(
			message.MAuditLogRecovering,
			"recovering previously aborted upload for audit log %s...", name,
		).Label("log", name),
	)
	file := path.Join(q.directory, name)

	// We are deliberately opening a file here.
	readHandle, err := os.Open(file) //nolint:gosec
	if err != nil {
		return err
	}

	// Remove existing multipart upload
	if err = q.abortMultiPartUpload(name); err != nil {
		return err
	}
	metadata := &queueEntryMetadata{}
	q.readMetadataFile(name, file, metadata)

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

func (q *uploadQueue) readMetadataFile(name string, file string, metadata *queueEntryMetadata) {
	metadataHandle, err := os.Open(fmt.Sprintf("%s.metadata.json", file))
	if err == nil {
		readBytes, err := ioutil.ReadAll(metadataHandle)
		if err != nil {
			q.logger.Error(
				message.Wrap(
					err,
					message.EAuditLogFailedReadingMetadataFile,
					"failed to read audit log %s metadata file",
					name,
				).Label("log", name),
			)
		} else {
			if err := json.Unmarshal(readBytes, metadata); err != nil {
				q.logger.Error(
					message.Wrap(
						err,
						message.EAuditLogFailedReadingMetadataFile,
						"failed to unmarshal audit log %s JSON",
						name,
					).Label("log", name),
				)
			}
		}
		if err := metadataHandle.Close(); err != nil {
			q.logger.Error(
				message.Wrap(
					err,
					message.EAuditLogFailedReadingMetadataFile,
					"failed to close audit log %s metadata file",
					name,
				).Label("log", name),
			)
		}
	} else {
		q.logger.Error(
			message.Wrap(
				err,
				message.EAuditLogFailedReadingMetadataFile,
				"metadata file for recovered audit log %s failed to open",
				name,
			).Label("log", name),
		)
	}
}
