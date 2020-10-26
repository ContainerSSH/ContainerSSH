package audit

type MessageType int32

const (
	MessageType_Connect                    MessageType = 0
	MessageType_Disconnect                 MessageType = 1
	MessageType_AuthPassword               MessageType = 100
	MessageType_AuthPasswordSuccessful     MessageType = 101
	MessageType_AuthPasswordFailed         MessageType = 102
	MessageType_AuthPasswordBackendError   MessageType = 103
	MessageType_AuthPubKey                 MessageType = 104
	MessageType_AuthPubKeySuccessful       MessageType = 105
	MessageType_AuthPubKeyFailed           MessageType = 106
	MessageType_AuthPubKeyBackendError     MessageType = 107
	MessageType_GlobalRequestUnknown       MessageType = 200
	MessageType_NewChannel                 MessageType = 300
	MessageType_NewChannelSuccessful       MessageType = 301
	MessageType_NewChannelFailed           MessageType = 302
	MessageType_ChannelRequestUnknownType  MessageType = 400
	MessageType_ChannelRequestDecodeFailed MessageType = 401
	MessageType_ChannelRequestSetEnv       MessageType = 402
	MessageType_ChannelRequestExec         MessageType = 403
	MessageType_ChannelRequestPty          MessageType = 404
	MessageType_ChannelRequestShell        MessageType = 405
	MessageType_ChannelRequestSignal       MessageType = 406
	MessageType_ChannelRequestSubsystem    MessageType = 407
	MessageType_ChannelRequestWindow       MessageType = 408
	MessageType_IO                         MessageType = 500
)
