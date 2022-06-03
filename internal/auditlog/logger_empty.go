package auditlog

import (
	"context"
	"io"
	"net"

	"github.com/containerssh/libcontainerssh/auditlog/message"
)

type empty struct{}

func (e *empty) OnAuthKeyboardInteractiveChallenge(
	_ string,
	_ string,
	_ []message.KeyboardInteractiveQuestion,
) {
}

func (e *empty) OnAuthKeyboardInteractiveAnswer(_ string, _ []message.KeyboardInteractiveAnswer) {}

func (e *empty) OnAuthKeyboardInteractiveFailed(_ string) {}

func (e *empty) OnAuthKeyboardInteractiveBackendError(_ string, _ string) {}

func (e *empty) OnRequestUnknown(_ uint64, _ string, _ []byte) {}

func (e *empty) OnRequestDecodeFailed(_ uint64, _ string, _ []byte, _ string) {}

func (e *empty) OnRequestFailed(_ uint64, _ error) {}

func (e *empty) OnRequestSetEnv(_ uint64, _ string, _ string) {}

func (e *empty) OnRequestExec(_ uint64, _ string) {}

func (e *empty) OnRequestPty(_ uint64, _ string, _ uint32, _ uint32, _ uint32, _ uint32, _ []byte) {}

func (e *empty) OnRequestX11(_ uint64, _ bool, _ string, _ string, _ uint32) {}

func (e *empty) OnRequestShell(_ uint64) {}

func (e *empty) OnRequestSignal(_ uint64, _ string) {}

func (e *empty) OnRequestSubsystem(_ uint64, _ string) {}

func (e *empty) OnRequestWindow(_ uint64, _ uint32, _ uint32, _ uint32, _ uint32) {}

func (e *empty) GetStdinProxy(reader io.Reader) io.Reader {
	return reader
}

func (e *empty) GetStdoutProxy(writer io.Writer) io.Writer {
	return writer
}

func (e *empty) GetStderrProxy(writer io.Writer) io.Writer {
	return writer
}

func (l *empty) GetForwardingProxy(forward io.ReadWriteCloser) io.ReadWriteCloser {
	return forward
}

func (e *empty) OnExit(_ uint32) {}

func (e *empty) OnExitSignal(_ string, _ bool, _ string, _ string) {}

func (e *empty) OnWriteClose() {}

func (e *empty) OnClose() {}

func (e *empty) OnDisconnect() {}

func (e *empty) OnAuthPassword(_ string, _ []byte) {}

func (e *empty) OnAuthPasswordSuccess(_ string, _ []byte) {}

func (e *empty) OnAuthPasswordFailed(_ string, _ []byte) {}

func (e *empty) OnAuthPasswordBackendError(_ string, _ []byte, _ string) {}

func (e *empty) OnAuthPubKey(_ string, _ string) {}

func (e *empty) OnAuthPubKeySuccess(_ string, _ string) {}

func (e *empty) OnAuthPubKeyFailed(_ string, _ string) {}

func (e *empty) OnAuthPubKeyBackendError(_ string, _ string, _ string) {}

func (e *empty) OnHandshakeFailed(_ string) {}

func (e *empty) OnHandshakeSuccessful(_ string) {}

func (e *empty) OnGlobalRequestUnknown(_ string) {}

func (e *empty) OnGlobalRequestDecodeFailed(_ uint64, _ string, _ []byte, _ error) {}

func (e *empty) OnNewChannel(_ message.ChannelID, _ string) {}

func (e *empty) OnNewChannelFailed(_ message.ChannelID, _ string, _ string) {}

func (e *empty) OnNewChannelSuccess(_ message.ChannelID, _ string) Channel {
	return e
}

func (l *empty) OnRequestTCPReverseForward(_ string, _ uint32) {}

func (l *empty) OnRequestCancelTCPReverseForward(_ string, _ uint32) {}

func (l *empty) OnTCPForwardChannel(_ message.ChannelID, _ string, _ uint32, _ string, _ uint32) {}

func (l *empty) OnReverseForwardChannel(_ message.ChannelID, _ string, _ uint32, _ string, _ uint32) {
}

func (l *empty) OnReverseStreamLocalChannel(_ message.ChannelID, _ string) {}

func (l *empty) OnReverseX11ForwardChannel(_ message.ChannelID, _ string, _ uint32) {}

func (l *empty) OnDirectStreamLocal(_ message.ChannelID, _ string) {}

func (l *empty) OnRequestStreamLocal(_ string) {}

func (l *empty) OnRequestCancelStreamLocal(_ string) {}

func (e *empty) OnConnect(_ message.ConnectionID, _ net.TCPAddr) (Connection, error) {
	return e, nil
}

func (e *empty) Shutdown(_ context.Context) {}
