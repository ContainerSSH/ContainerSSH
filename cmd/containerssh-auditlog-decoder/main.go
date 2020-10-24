package main

import (
	"compress/gzip"
	"github.com/containerssh/containerssh/audit/protocol"
	"github.com/fxamacker/cbor"
	"github.com/mitchellh/mapstructure"
	"log"
	"os"
)

func main() {
	auditFile, err := os.Open(os.Args[1])

	if err != nil {
		log.Fatalf("failed to open audit log file %s (%v)", os.Args[1], err)
	}

	gzipReader, err := gzip.NewReader(auditFile)

	if err != nil {
		log.Fatalf("failed to open gzip stream of file %s (%v)", os.Args[1], err)
	}

	cborReader := cbor.NewDecoder(gzipReader)

	messages := []protocol.Message{}

	err = cborReader.Decode(&messages)

	if err != nil {
		log.Fatalf("failed to decode messages (%v)", err)
	}

	for _, v := range messages {
		var payload interface{}

		switch v.MessageType {
		/*
			The following message types have no payload: MessageType_Disconnect, MessageType_AuthPasswordSuccessful, MessageType_AuthPasswordFailed, MessageType_AuthPasswordBackendError, MessageType_AuthPubKeySuccessful, MessageType_AuthPubKeyFailed, MessageType_AuthPubKeyBackendError,
		*/

		case protocol.MessageType_Connect:
			payload = &protocol.MessageConnect{}
		case protocol.MessageType_AuthPassword:
			payload = &protocol.MessageAuthPassword{}
		case protocol.MessageType_AuthPubKey:
			payload = &protocol.MessageAuthPubKey{}
		case protocol.MessageType_GlobalRequestUnknown:
			payload = &protocol.MessageGlobalRequestUnknown{}
		case protocol.MessageType_UnknownChannelType:
			payload = &protocol.MessageUnknownChannelType{}
		case protocol.MessageType_NewChannel:
			payload = &protocol.MessageNewChannel{}
		case protocol.MessageType_UnknownChannelRequestType:
			payload = &protocol.MessageUnknownChannelRequestType{}
		case protocol.MessageType_FailedToDecodeChannelRequest:
			payload = &protocol.MessageFailedToDecodeChannelRequest{}
		case protocol.MessageType_ChannelRequestSetEnv:
			payload = &protocol.MessageChannelRequestSetEnv{}
		case protocol.MessageType_ChannelRequestExec:
			payload = &protocol.MessageChannelRequestExec{}
		case protocol.MessageType_ChannelRequestPty:
			payload = &protocol.MessageChannelRequestPty{}
		case protocol.MessageType_ChannelRequestShell:
			payload = &protocol.MessageChannelRequestShell{}
		case protocol.MessageType_ChannelRequestSignal:
			payload = &protocol.MessageChannelRequestSignal{}
		case protocol.MessageType_ChannelRequestSubsystem:
			payload = &protocol.MessageChannelRequestSubsystem{}
		case protocol.MessageType_ChannelRequestWindow:
			payload = &protocol.MessageChannelRequestWindow{}
		case protocol.MessageType_IO:
			payload = &protocol.MessageIO{}

		}
		if payload != nil {
			err = mapstructure.Decode(v.Payload, payload)
			if err != nil {
				log.Fatalf("failed to decode payload (%v)", err)
			}
		}
	}
}
