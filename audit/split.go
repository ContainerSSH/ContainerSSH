package audit

import (
	"encoding/hex"
	"io"
	"sync"

	"github.com/containerssh/containerssh/audit/format/audit"
	"github.com/containerssh/containerssh/log"
)

func NewPlugin(logger log.Logger, storage Storage, encoder Encoder) Plugin {
	return &SplitAuditPlugin{
		connections: sync.Map{},
		logger:      logger,
		storage:     storage,
		encoder:     encoder,
	}
}

type SplitAuditPlugin struct {
	connections sync.Map
	logger      log.Logger
	storage     Storage
	encoder     Encoder
}

type connectionEntry struct {
	writer         io.WriteCloser
	messageChannel chan audit.Message
}

func (p *SplitAuditPlugin) Message(msg audit.Message) {
	name := hex.EncodeToString(msg.ConnectionID)
	var entry connectionEntry
	e, ok := p.connections.Load(name)
	if !ok {
		writer, err := p.storage.Open(name)
		if err != nil {
			p.logger.WarningF("failed to open storage (%v)", err)
			return
		}
		entry = connectionEntry{
			writer:         writer,
			messageChannel: make(chan audit.Message),
		}
		p.connections.Store(name, entry)
		go func() {
			p.encoder.Encode(entry.messageChannel, entry.writer)
			err := entry.writer.Close()
			if err != nil {
				p.logger.WarningF("failed to close audit log file %s (%v)", name, err)
			}
			p.connections.Delete(name)
		}()
	} else {
		entry = e.(connectionEntry)
	}
	entry.messageChannel <- msg
	if msg.MessageType == audit.MessageType_Disconnect {
		close(entry.messageChannel)
	}
}
