package asciinema

import (
	"encoding/json"
	"fmt"
	"io"
	"net"

    "go.containerssh.io/libcontainerssh/auditlog/message"
    "go.containerssh.io/libcontainerssh/internal/auditlog/storage"
    "go.containerssh.io/libcontainerssh/internal/geoip/geoipprovider"
    "go.containerssh.io/libcontainerssh/log"
    messageCodes "go.containerssh.io/libcontainerssh/message"
)

type encoder struct {
	logger        log.Logger
	geoIPProvider geoipprovider.LookupProvider
}

func (e *encoder) GetMimeType() string {
	return "application/x-asciicast"
}

func (e *encoder) GetFileExtension() string {
	return ".cast"
}

func (e *encoder) sendHeader(header Header, storage io.Writer) error {
	data, err := json.Marshal(header)
	if err != nil {
		return err
	}
	_, err = storage.Write(append(data, '\n'))
	if err != nil {
		return err
	}
	return nil
}

func (e *encoder) sendFrame(frame Frame, storage io.Writer) error {
	data, err := frame.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal Asciicast frame (%w)", err)
	}
	if _, err = storage.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write Asciicast frame (%w)", err)
	}
	return nil
}

func (e *encoder) Encode(messages <-chan message.Message, storage storage.Writer) error {
	asciicastHeader := Header{
		Version:   2,
		Width:     80,
		Height:    25,
		Timestamp: 0,
		Command:   "",
		Title:     "",
		Env:       map[string]string{},
	}
	startTime := int64(0)
	headerWritten := false
	var ip = ""
	var username *string
	const shell = "/bin/sh"
	for {
		msg, ok := <-messages
		if !ok {
			break
		}
		var err error
		startTime, headerWritten, ip, username, err = e.encodeMessage(
			startTime,
			msg,
			&asciicastHeader,
			ip,
			storage,
			username,
			headerWritten,
			shell,
		)
		if err != nil {
			if err := storage.Close(); err != nil {
				e.logger.Error(messageCodes.Wrap(err, messageCodes.EAuditLogStorageCloseFailed, "failed to close audit log storage writer"))
			}
			return err
		}
	}
	if !headerWritten {
		if err := e.sendHeader(asciicastHeader, storage); err != nil {
			if err := storage.Close(); err != nil {
				e.logger.Error(messageCodes.Wrap(err, messageCodes.EAuditLogStorageCloseFailed, "failed to close audit log storage writer"))
			}
			return err
		}
	}
	if err := storage.Close(); err != nil {
		e.logger.Error(messageCodes.Wrap(err, messageCodes.EAuditLogStorageCloseFailed, "failed to close audit log storage writer"))
	}
	return nil
}

func (e *encoder) encodeMessage(
	startTime int64,
	msg message.Message,
	asciicastHeader *Header,
	ip string,
	storage storage.Writer,
	username *string,
	headerWritten bool,
	shell string,
) (int64, bool, string, *string, error) {
	if msg.MessageType == message.TypeConnect {
		startTime = msg.Timestamp
		asciicastHeader.Timestamp = int(startTime / 1000000000)
	}
	var err error
	country := e.geoIPProvider.Lookup(net.ParseIP(ip))
	switch msg.MessageType {
	case message.TypeConnect:
		ip, username = e.handleConnect(storage, msg, startTime, country, username)
	case message.TypeAuthPasswordSuccessful:
		ip, username = e.handleAuthPasswordSuccessful(storage, msg, startTime, ip, country)
	case message.TypeAuthPubKeySuccessful:
		ip, username = e.handleAuthPubkeySuccessful(storage, msg, startTime, ip, country)
	case message.TypeHandshakeSuccessful:
		ip, username = e.handleHandshakeSuccessful(storage, msg, startTime, ip, country)
	case message.TypeChannelRequestSetEnv:
		payload := msg.Payload.(message.PayloadChannelRequestSetEnv)
		asciicastHeader.Env[payload.Name] = payload.Value
	case message.TypeChannelRequestPty:
		e.handleChannelRequestPty(msg, asciicastHeader)
	case message.TypeChannelRequestExec:
		payload := msg.Payload.(message.PayloadChannelRequestExec)
		startTime, headerWritten, err = e.handleRun(startTime, headerWritten, asciicastHeader, payload.Program, storage)
	case message.TypeChannelRequestShell:
		startTime, headerWritten, err = e.handleRun(startTime, headerWritten, asciicastHeader, shell, storage)
	case message.TypeChannelRequestSubsystem:
		startTime, headerWritten, err = e.handleRun(startTime, headerWritten, asciicastHeader, shell, storage)
	case message.TypeIO:
		startTime, headerWritten, err = e.handleIO(startTime, msg, asciicastHeader, headerWritten, shell, storage)
	}
	if err != nil {
		return startTime, headerWritten, ip, username, err
	}
	return startTime, headerWritten, ip, username, nil
}

func (e *encoder) handleConnect(storage storage.Writer, msg message.Message, startTime int64, country string, username *string) (string, *string) {
	payload := msg.Payload.(message.PayloadConnect)
	ip := payload.RemoteAddr
	storage.SetMetadata(startTime/1000000000, ip, country, username)
	return ip, username
}

func (e *encoder) handleAuthPasswordSuccessful(storage storage.Writer, msg message.Message, startTime int64, ip string, country string) (string, *string) {
	payload := msg.Payload.(message.PayloadAuthPassword)
	username := &payload.Username
	storage.SetMetadata(startTime/1000000000, ip, country, username)
	return ip, username
}

func (e *encoder) handleAuthPubkeySuccessful(storage storage.Writer, msg message.Message, startTime int64, ip string, country string) (string, *string) {
	payload := msg.Payload.(message.PayloadAuthPubKey)
	username := &payload.Username
	storage.SetMetadata(startTime/1000000000, ip, country, username)
	return ip, username
}

func (e *encoder) handleHandshakeSuccessful(storage storage.Writer, msg message.Message, startTime int64, ip string, country string) (string, *string) {
	payload := msg.Payload.(message.PayloadHandshakeSuccessful)
	username := &payload.Username
	storage.SetMetadata(startTime/1000000000, ip, country, username)
	return ip, username
}

func (e *encoder) handleChannelRequestPty(msg message.Message, asciicastHeader *Header) {
	payload := msg.Payload.(message.PayloadChannelRequestPty)
	asciicastHeader.Env["TERM"] = payload.Term
	asciicastHeader.Width = uint(payload.Columns)
	asciicastHeader.Height = uint(payload.Rows)
}

func (e *encoder) handleRun(startTime int64, headerWritten bool, asciicastHeader *Header, program string, storage storage.Writer) (int64, bool, error) {
	if !headerWritten {
		asciicastHeader.Command = program
		if err := e.sendHeader(*asciicastHeader, storage); err != nil {
			return startTime, headerWritten, err
		}
		headerWritten = true
	}
	return startTime, headerWritten, nil
}

func (e *encoder) handleIO(startTime int64, msg message.Message, asciicastHeader *Header, headerWritten bool, shell string, storage storage.Writer) (int64, bool, error) {
	if !headerWritten {
		asciicastHeader.Command = shell
		if err := e.sendHeader(*asciicastHeader, storage); err != nil {
			return startTime, headerWritten, err
		}
		headerWritten = true
	}
	payload := msg.Payload.(message.PayloadIO)
	if payload.Stream == message.StreamStdout ||
		payload.Stream == message.StreamStderr {
		time := float64(msg.Timestamp-startTime) / 1000000000
		frame := Frame{
			Time:      time,
			EventType: EventTypeOutput,
			Data:      string(payload.Data),
		}
		if err := e.sendFrame(frame, storage); err != nil {
			return startTime, headerWritten, err
		}
	}
	return startTime, headerWritten, nil
}
