package audit

import (
	"compress/gzip"
	"github.com/containerssh/containerssh/log"
	"github.com/fxamacker/cbor"
	"io"
)

type Encoder struct {
	logger log.Logger
}

func NewEncoder(logger log.Logger) (*Encoder, error) {
	return &Encoder{
		logger: logger,
	}, nil
}

func (e *Encoder) Encode(messages <-chan Message, storage io.Writer) {
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

	for {
		msg, ok := <-messages
		if !ok {
			break
		}
		err = encoder.Encode(&msg)
		if err != nil {
			e.logger.ErrorF("failed to encode audit log message (%v)", err)
		}
		if msg.MessageType == MessageType_Disconnect {
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
