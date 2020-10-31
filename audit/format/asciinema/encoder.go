package asciinema

import (
	"encoding/json"
	"github.com/containerssh/containerssh/audit"
	auditFormat "github.com/containerssh/containerssh/audit/format/audit"
	"github.com/containerssh/containerssh/log"
	"io"
)

type encoder struct {
	logger log.Logger
}

func (e *encoder) sendHeader(header AsciicastHeader, storage io.Writer) {
	data, err := json.Marshal(header)
	if err != nil {
		e.logger.WarningF("failed to marshal Asciicast header (%v)", err)
	} else {
		_, err = storage.Write(append(data, '\n'))
		if err != nil {
			e.logger.WarningF("failed to write Asciicast header (%v)", err)
		}
	}
}

func (e *encoder) sendFrame(frame AsciicastFrame, storage io.Writer) {
	data, err := frame.MarshalJSON()
	if err != nil {
		e.logger.WarningF("failed to marshal Asciicast frame (%v)", err)
	} else {
		_, err = storage.Write(append(data, '\n'))
		if err != nil {
			e.logger.WarningF("failed to write Asciicast frame (%v)", err)
		}
	}
}

func (e *encoder) Encode(messages <-chan auditFormat.Message, storage audit.StorageWriter) {
	asciicastHeader := AsciicastHeader{
		Version:   2,
		Width:     80,
		Height:    25,
		Timestamp: 0,
		Command:   "",
		Title:     "",
		Env:       map[string]string{},
	}
	startTime := int64(0)
	headerSent := false
	var ip = ""
	var username *string
	for {
		msg, ok := <-messages
		if !ok {
			break
		}
		if startTime == 0 {
			startTime = msg.Timestamp
			asciicastHeader.Timestamp = int(startTime / 1000000000)
		}
		switch msg.MessageType {
		case auditFormat.MessageType_Connect:
			payload := msg.Payload.(*auditFormat.PayloadConnect)
			ip = payload.RemoteAddr
			storage.SetMetadata(startTime/1000000000, ip, username)
		case auditFormat.MessageType_AuthPasswordSuccessful:
			payload := msg.Payload.(*auditFormat.PayloadAuthPassword)
			username = &payload.Username
			storage.SetMetadata(startTime/1000000000, ip, username)
		case auditFormat.MessageType_AuthPubKeySuccessful:
			payload := msg.Payload.(*auditFormat.PayloadAuthPubKey)
			username = &payload.Username
			storage.SetMetadata(startTime/1000000000, ip, username)
		case auditFormat.MessageType_ChannelRequestSetEnv:
			if headerSent {
				break
			}
			payload := msg.Payload.(*auditFormat.PayloadChannelRequestSetEnv)
			asciicastHeader.Env[payload.Name] = payload.Value
		case auditFormat.MessageType_ChannelRequestPty:
			if headerSent {
				break
			}
			payload := msg.Payload.(*auditFormat.PayloadChannelRequestPty)
			asciicastHeader.Width = payload.Columns
			asciicastHeader.Height = payload.Rows
		case auditFormat.MessageType_ChannelRequestExec:
			if headerSent {
				break
			}
			payload := msg.Payload.(*auditFormat.PayloadChannelRequestExec)
			asciicastHeader.Command = payload.Program
			e.sendHeader(asciicastHeader, storage)
			headerSent = true
		case auditFormat.MessageType_ChannelRequestShell:
			if headerSent {
				break
			}
			asciicastHeader.Command = "/bin/sh"
			e.sendHeader(asciicastHeader, storage)
			headerSent = true
		case auditFormat.MessageType_ChannelRequestSubsystem:
			//Fallback
			if headerSent {
				break
			}
			asciicastHeader.Command = "/bin/sh"
			e.sendHeader(asciicastHeader, storage)
			headerSent = true
		case auditFormat.MessageType_IO:
			if !headerSent {
				asciicastHeader.Command = "/bin/sh"
				e.sendHeader(asciicastHeader, storage)
				headerSent = true
			}
			payload := msg.Payload.(*auditFormat.PayloadIO)
			if payload.Stream == auditFormat.Stream_StdOut ||
				payload.Stream == auditFormat.Stream_StdErr {
				time := float64(msg.Timestamp-startTime) / 1000000000
				frame := AsciicastFrame{
					Time:      time,
					EventType: AsciicastEventTypeOutput,
					Data:      string(payload.Data),
				}
				e.sendFrame(frame, storage)
			}
		}
	}

}
