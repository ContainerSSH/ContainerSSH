package protocol

import (
	"net"
)

type MessageType int

const (
	MessageType_Connect                      MessageType = 0
	MessageType_Disconnect                   MessageType = 1
	MessageType_AuthPassword                 MessageType = 100
	MessageType_AuthPasswordSuccessful       MessageType = 101
	MessageType_AuthPasswordFailed           MessageType = 102
	MessageType_AuthPasswordBackendError     MessageType = 103
	MessageType_AuthPubKey                   MessageType = 104
	MessageType_AuthPubKeySuccessful         MessageType = 105
	MessageType_AuthPubKeyFailed             MessageType = 106
	MessageType_AuthPubKeyBackendError       MessageType = 107
	MessageType_GlobalRequestUnknown         MessageType = 200
	MessageType_UnknownChannelType           MessageType = 300
	MessageType_NewChannel                   MessageType = 301
	MessageType_UnknownChannelRequestType    MessageType = 302
	MessageType_FailedToDecodeChannelRequest MessageType = 303
	MessageType_ChannelRequestSetEnv         MessageType = 400
	MessageType_ChannelRequestExec           MessageType = 401
	MessageType_ChannelRequestPty            MessageType = 402
	MessageType_ChannelRequestShell          MessageType = 403
	MessageType_ChannelRequestSignal         MessageType = 404
	MessageType_ChannelRequestSubsystem      MessageType = 405
	MessageType_ChannelRequestWindow         MessageType = 406
	MessageType_IO                           MessageType = 500
)

type ChannelId int64

type Message struct {
	ConnectionID []byte      `json:"connectionId" yaml:"connectionId"`
	Timestamp    int64       `json:"timestamp" yaml:"timestamp"`
	MessageType  MessageType `json:"type" yaml:"type"`
	Payload      interface{} `json:"payload" yaml:"payload"`
	ChannelId    ChannelId
}

type MessageConnect struct {
	RemoteAddr net.IP `json:"remoteAddr" yaml:"remoteAddr"`
}

type MessageAuthPassword struct {
	Username string `json:"username" yaml:"username"`
	Password []byte `json:"password" yaml:"password"`
}

type MessageAuthPubKey struct {
	Username string `json:"username" yaml:"username"`
	Key      []byte `json:"key" yaml:"key"`
}

type MessageGlobalRequestUnknown struct {
	RequestType string `json:"requestType" yaml:"requestType"`
}

type MessageUnknownChannelType struct {
	ChannelType string `json:"channelType" yaml:"channelType"`
}

type MessageNewChannel struct {
	ChannelType string `json:"channelType" yaml:"channelType"`
}

type MessageUnknownChannelRequestType struct {
	RequestType string `json:"requestType" yaml:"requestType"`
}

type MessageFailedToDecodeChannelRequest struct {
	RequestType string `json:"requestType" yaml:"requestType"`
}

type MessageChannelRequestSetEnv struct {
	Name  string `json:"name" yaml:"name"`
	Value string `json:"value" yaml:"value"`
}

type MessageChannelRequestExec struct {
	Program string `json:"program" yaml:"program"`
}

type MessageChannelRequestPty struct {
	Columns uint `json:"columns" yaml:"columns"`
	Rows    uint `json:"rows" yaml:"rows"`
}

type MessageChannelRequestShell struct {
}

type MessageChannelRequestSignal struct {
	Signal string `json:"signal" yaml:"signal"`
}

type MessageChannelRequestSubsystem struct {
	Subsystem string `json:"subsystem" yaml:"subsystem"`
}

type MessageChannelRequestWindow struct {
	Columns uint `json:"columns" yaml:"columns"`
	Rows    uint `json:"rows" yaml:"rows"`
}

type Stream uint

const (
	Stream_Stdin  Stream = 0
	Stream_StdOut Stream = 1
	Stream_StdErr Stream = 2
)

type MessageIO struct {
	Stream Stream `json:"stream" yaml:"stream"`
	Data   []byte `json:"data" yaml:"data"`
}
