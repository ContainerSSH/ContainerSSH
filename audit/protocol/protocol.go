package protocol

import "net"

type MessageType int

const (
	MessageType_Connect                  MessageType = 0
	MessageType_Disconnect               MessageType = 1
	MessageType_AuthPassword             MessageType = 10
	MessageType_AuthPasswordSuccessful   MessageType = 11
	MessageType_AuthPasswordFailed       MessageType = 12
	MessageType_AuthPasswordBackendError MessageType = 13
	MessageType_AuthPubKey               MessageType = 14
	MessageType_AuthPubKeySuccessful     MessageType = 15
	MessageType_AuthPubKeyFailed         MessageType = 16
	MessageType_AuthPubKeyBackendError   MessageType = 17
)

type MessageConnectionID [16]byte

type Message struct {
	ConnectionID MessageConnectionID
	Timestamp    int64
	MessageType  MessageType
	Payload      interface{}
}

type MessageConnectPayload struct {
	RemoteAddr net.IP
}

type MessageAuthPayloadPassword struct {
	Username string
	Password []byte
}

type MessageAuthPayloadPubKey struct {
	Username string
	Key      []byte
}

