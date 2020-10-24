package file

import (
	"compress/gzip"
	"encoding/hex"
	"os"
	"path"
	"sync"

	"github.com/containerssh/containerssh/audit/format"
	"github.com/containerssh/containerssh/log"

	"github.com/fxamacker/cbor"
)

type auditLogFile struct {
	fileHandle *os.File
	gzipHandle *gzip.Writer
	encoder    *cbor.Encoder
}

type Plugin struct {
	directory   string
	connections sync.Map
	logger      log.Logger
}

func (p *Plugin) Message(msg format.Message) {
	fileName := hex.EncodeToString(msg.ConnectionID)
	auditLogEntry, ok := p.connections.Load(fileName)
	var fileHandle *os.File
	var gzipHandle *gzip.Writer
	var err error
	var encoder *cbor.Encoder
	if !ok {
		fileHandle, err = os.Create(path.Join(p.directory, fileName))
		if err != nil {
			p.logger.ErrorF("failed to open audit log file %s (%v)", fileName, err)
			return
		}
		gzipHandle = gzip.NewWriter(fileHandle)
		encoder = cbor.NewEncoder(gzipHandle, cbor.EncOptions{})
		err = encoder.StartIndefiniteArray()
		if err != nil {
			p.logger.ErrorF("failed to start infinite array in audit log file %s (%v)", fileName, err)
			return
		}
		p.connections.Store(fileName, auditLogFile{
			fileHandle: fileHandle,
			gzipHandle: gzipHandle,
			encoder:    encoder,
		})
	} else {
		fileHandle = auditLogEntry.(auditLogFile).fileHandle
		gzipHandle = auditLogEntry.(auditLogFile).gzipHandle
		encoder = auditLogEntry.(auditLogFile).encoder
	}

	err = encoder.Encode(&msg)
	if err != nil {
		p.logger.ErrorF("failed to encode audit log message (%v)", err)
	}

	if msg.MessageType == format.MessageType_Disconnect {
		defer p.connections.Delete(fileName)
		err := encoder.EndIndefinite()
		if err != nil {
			p.logger.WarningF("failed to end audit log infinite array (%v)", err)
			return
		}

		err = gzipHandle.Flush()
		if err != nil {
			p.logger.WarningF("failed to flush audit log gzip stream (%v)", err)
			return
		}
		err = gzipHandle.Close()
		if err != nil {
			p.logger.WarningF("failed to close audit log gzip stream (%v)", err)
			return
		}
		err = fileHandle.Close()
		if err != nil {
			p.logger.WarningF("failed to close audit log file %s (%v)", fileName, err)
			return
		}
	}
}
