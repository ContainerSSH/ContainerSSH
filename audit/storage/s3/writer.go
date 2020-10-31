package s3

import (
	"github.com/containerssh/containerssh/audit"
	"io"
)

func newMonitoringWriter(
	backingWriter io.WriteCloser,
	partSize uint,
	onMetadata func(startTime int64, remoteAddr string, username *string),
	onPart func(),
	onClose func(),
) audit.StorageWriter {
	return &monitoringWriter{
		backingWriter: backingWriter,
		bytesWritten:  0,
		partSize:      partSize,
		lastPart:      0,
		onMetadata:    onMetadata,
		onPart:        onPart,
		onClose:       onClose,
	}
}

// The monitoring writer writes to a backing writer and notifies a configured callback when a new part of a given size
// is available, or when the writer is closed.
type monitoringWriter struct {
	backingWriter io.WriteCloser
	bytesWritten  uint64
	closed        bool
	partSize      uint
	onMetadata    func(startTime int64, remoteAddr string, username *string)
	onPart        func()
	onClose       func()
	lastPart      int
}

func (m *monitoringWriter) SetMetadata(startTime int64, sourceIp string, username *string) {
	m.onMetadata(startTime, sourceIp, username)
}

func (m *monitoringWriter) Write(p []byte) (n int, err error) {
	bytes, err := m.backingWriter.Write(p)
	m.bytesWritten += uint64(bytes)
	partsAvailable := int(m.bytesWritten / uint64(m.partSize))
	if partsAvailable > m.lastPart {
		go m.onPart()
	}
	return bytes, err
}

func (m *monitoringWriter) Close() error {
	err := m.backingWriter.Close()
	go m.onClose()
	return err
}
