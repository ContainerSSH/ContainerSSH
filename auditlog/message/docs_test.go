package message_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.containerssh.io/containerssh/auditlog/message"
)

func TestDocumentMessages(t *testing.T) {
	docs := message.DocumentMessages()

	var keys []string
	for _, field := range docs.Message.Fields {
		keys = append(keys, field.Name)
	}
	// The binary audit log uses these names as CBOR map keys.
	assert.Equal(t, []string{"connectionId", "timestamp", "type", "payload", "channelId"}, keys)

	var codes []int32
	for docType := range docs.Payloads {
		codes = append(codes, docType.Code())
		assert.NotEqual(t, "invalid", docType.Name(), "message type %d has no name", docType.Code())
	}
	assert.Contains(t, codes, message.TypeHandshakeFailed.Code())
	assert.Contains(t, codes, message.TypeHandshakeSuccessful.Code())
}
