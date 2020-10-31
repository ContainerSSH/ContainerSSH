package format

import (
	"compress/gzip"
	"github.com/containerssh/containerssh/audit"
	audit2 "github.com/containerssh/containerssh/audit/format/audit"
	"github.com/containerssh/containerssh/log"
	"github.com/fxamacker/cbor"
)

type Encoder struct {
	logger log.Logger
}

func NewEncoder(logger log.Logger) (*Encoder, error) {
	return &Encoder{
		logger: logger,
	}, nil
}

func (e *Encoder) Encode(messages <-chan audit2.Message, storage audit.StorageWriter) {
	var gzipHandle *gzip.Writer
	var err error
	var encoder *cbor.Encoder
	gzipHandle = gzip.NewWriter(storage)
	encoder = cbor.NewEncoder(gzipHandle, cbor.EncOptions{})
	err = encoder.StartIndefiniteArray()
	if err != nil {
		e.logger.ErrorF("failed to start infinite array (%v)", err)
		return
	}

	startTime := int64(0)
	var ip = ""
	var username *string
	for {
		msg, ok := <-messages
		if !ok {
			break
		}
		if startTime == 0 {
			startTime = msg.Timestamp
		}
		switch msg.MessageType {
		case audit2.MessageType_Connect:
			payload := msg.Payload.(*audit2.PayloadConnect)
			ip = payload.RemoteAddr
			storage.SetMetadata(startTime/1000000000, ip, username)
		case audit2.MessageType_AuthPasswordSuccessful:
			payload := msg.Payload.(*audit2.PayloadAuthPassword)
			username = &payload.Username
			storage.SetMetadata(startTime/1000000000, ip, username)
		case audit2.MessageType_AuthPubKeySuccessful:
			payload := msg.Payload.(*audit2.PayloadAuthPassword)
			username = &payload.Username
			storage.SetMetadata(startTime/1000000000, ip, username)
		}
		err = encoder.Encode(&msg)
		if err != nil {
			e.logger.ErrorF("failed to encode audit log message (%v)", err)
		}
		if msg.MessageType == audit2.MessageType_Disconnect {
			break
		}
	}
	err = encoder.EndIndefinite()
	if err != nil {
		e.logger.WarningF("failed to end audit log infinite array (%v)", err)
		return
	}

	err = gzipHandle.Flush()
	if err != nil {
		e.logger.WarningF("failed to flush audit log gzip stream (%v)", err)
		return
	}
	err = gzipHandle.Close()
	if err != nil {
		e.logger.WarningF("failed to close audit log gzip stream (%v)", err)
		return
	}
}
