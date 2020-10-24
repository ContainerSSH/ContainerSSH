package format

import (
	"compress/gzip"
	"fmt"
	"github.com/fxamacker/cbor"
	"github.com/mitchellh/mapstructure"
	"io"
	"log"
)

type DecodedMessage struct {
	Message
	TypeName string `json:"typeName" yaml:"typeName"`
}

func Decode(reader io.Reader) (<-chan *DecodedMessage, <-chan error, <-chan bool) {
	result := make(chan *DecodedMessage)
	errors := make(chan error)
	done := make(chan bool, 1)

	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		defer func() {
			errors <- fmt.Errorf("failed to open gzip stream (%v)", err)
			done <- true
			close(result)
			close(errors)
			close(done)
		}()
		return result, errors, done
	}

	cborReader := cbor.NewDecoder(gzipReader)

	var messages []Message

	err = cborReader.Decode(&messages)
	if err != nil {
		defer func() {
			errors <- fmt.Errorf("failed to decode messages (%v)", err)
			done <- true
			close(result)
			close(errors)
			close(done)

		}()
		return result, errors, done
	}

	go func() {
		for _, v := range messages {
			var payload interface{}

			switch v.MessageType {

			case MessageType_Connect:
				payload = &PayloadConnect{}
			case MessageType_Disconnect:

			case MessageType_AuthPassword:
				payload = &PayloadAuthPassword{}
			case MessageType_AuthPasswordSuccessful:
			case MessageType_AuthPasswordFailed:
			case MessageType_AuthPasswordBackendError:

			case MessageType_AuthPubKey:
				payload = &PayloadAuthPubKey{}
			case MessageType_AuthPubKeySuccessful:
			case MessageType_AuthPubKeyFailed:
			case MessageType_AuthPubKeyBackendError:

			case MessageType_GlobalRequestUnknown:
				payload = &PayloadGlobalRequestUnknown{}
			case MessageType_NewChannel:
				payload = &PayloadNewChannel{}
			case MessageType_NewChannelSuccessful:
				payload = &PayloadNewChannelSuccessful{}
			case MessageType_NewChannelFailed:
				payload = &PayloadNewChannelFailed{}

			case MessageType_ChannelRequestUnknownType:
				payload = &PayloadChannelRequestUnknownType{}
			case MessageType_ChannelRequestDecodeFailed:
				payload = &PayloadChannelRequestDecodeFailed{}
			case MessageType_ChannelRequestSetEnv:
				payload = &PayloadChannelRequestSetEnv{}
			case MessageType_ChannelRequestExec:
				payload = &PayloadChannelRequestExec{}
			case MessageType_ChannelRequestPty:
				payload = &PayloadChannelRequestPty{}
			case MessageType_ChannelRequestShell:
				payload = &PayloadChannelRequestShell{}
			case MessageType_ChannelRequestSignal:
				payload = &PayloadChannelRequestSignal{}
			case MessageType_ChannelRequestSubsystem:
				payload = &PayloadChannelRequestSubsystem{}
			case MessageType_ChannelRequestWindow:
				payload = &PayloadChannelRequestWindow{}
			case MessageType_IO:
				payload = &MessageIO{}
			default:
				errors <- fmt.Errorf("invalid message type: %d", v.MessageType)
			}
			if payload != nil {
				err = mapstructure.Decode(v.Payload, payload)
				if err != nil {
					log.Fatalf("failed to decode payload (%v)", err)
				}
				v.Payload = payload
			} else {
				v.Payload = nil
			}
			decodedMessage := &DecodedMessage{
				Message:  v,
				TypeName: TypeToName(v.MessageType),
			}
			result <- decodedMessage
		}
		done <- true
		close(result)
		close(errors)
		close(done)

	}()
	return result, errors, done
}

func TypeToName(messageType MessageType) string {
	switch messageType {

	case MessageType_Connect:
		return "auth_connect"
	case MessageType_Disconnect:
		return "disconnect"

	case MessageType_AuthPassword:
		return "auth_password"
	case MessageType_AuthPasswordSuccessful:
		return "auth_password_successful"
	case MessageType_AuthPasswordFailed:
		return "auth_password_failed"
	case MessageType_AuthPasswordBackendError:
		return "auth_password_backend_error"

	case MessageType_AuthPubKey:
		return "auth_pubkey"
	case MessageType_AuthPubKeySuccessful:
		return "auth_pubkey_successful"
	case MessageType_AuthPubKeyFailed:
		return "auth_pubkey_failed"
	case MessageType_AuthPubKeyBackendError:
		return "auth_pubkey_backend_error"

	case MessageType_GlobalRequestUnknown:
		return "global_request_unknown"
	case MessageType_NewChannel:
		return "new_channel"
	case MessageType_NewChannelSuccessful:
		return "new_channel_successful"
	case MessageType_NewChannelFailed:
		return "new_channel_failed"

	case MessageType_ChannelRequestUnknownType:
		return "channel_request_unknown"
	case MessageType_ChannelRequestDecodeFailed:
		return "channel_request_decode_failed"
	case MessageType_ChannelRequestSetEnv:
		return "setenv"
	case MessageType_ChannelRequestExec:
		return "exec"
	case MessageType_ChannelRequestPty:
		return "pty"
	case MessageType_ChannelRequestShell:
		return "shell"
	case MessageType_ChannelRequestSignal:
		return "signal"
	case MessageType_ChannelRequestSubsystem:
		return "subsystem"
	case MessageType_ChannelRequestWindow:
		return "window"
	case MessageType_IO:
		return "io"
	default:
		return "invalid"
	}
}
