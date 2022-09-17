package auditlogintegration_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.containerssh.io/libcontainerssh/auditlog/message"
	auth2 "go.containerssh.io/libcontainerssh/auth"
	"go.containerssh.io/libcontainerssh/config"
	"go.containerssh.io/libcontainerssh/internal/auditlog/codec/binary"
	"go.containerssh.io/libcontainerssh/internal/auditlog/storage/file"
	"go.containerssh.io/libcontainerssh/internal/auditlogintegration"
	"go.containerssh.io/libcontainerssh/internal/auth"
	"go.containerssh.io/libcontainerssh/internal/geoip"
	"go.containerssh.io/libcontainerssh/internal/sshserver"
	"go.containerssh.io/libcontainerssh/log"
	message2 "go.containerssh.io/libcontainerssh/message"
	"go.containerssh.io/libcontainerssh/metadata"
	"golang.org/x/crypto/ssh"
)

func TestKeyboardInteractiveAuthentication(t *testing.T) {
	logger := log.NewTestLogger(t)

	dir, err := os.MkdirTemp("temp", "testcase")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	geoipLookup, err := geoip.New(
		config.GeoIPConfig{
			Provider: "dummy",
		},
	)
	assert.NoError(t, err)

	auditLogHandler, err := auditlogintegration.New(
		config.AuditLogConfig{
			Enable:  true,
			Format:  config.AuditLogFormatBinary,
			Storage: config.AuditLogStorageFile,
			File: config.AuditLogFileConfig{
				Directory: dir,
			},
		},
		&backendHandler{},
		geoipLookup,
		logger,
	)
	assert.NoError(t, err)

	user := sshserver.NewTestUser("test")
	user.AddKeyboardInteractiveChallengeResponse("Challenge", "Response")

	srv := sshserver.NewTestServer(t, auditLogHandler, logger, nil)
	srv.Start()
	client := sshserver.NewTestClient(srv.GetListen(), srv.GetHostKey(), user, logger)
	connection := client.MustConnect()
	_ = connection.Close()
	srv.Stop(10 * time.Second)

	messages, errors, done := getStoredMessages(t, dir, logger)
	if done {
		return
	}
	assert.Empty(t, errors)
	assert.Equal(t, message.TypeConnect, messages[0].MessageType)
	assert.Equal(t, message.TypeAuthKeyboardInteractiveChallenge, messages[1].MessageType)
	assert.Equal(t, message.TypeAuthKeyboardInteractiveAnswer, messages[2].MessageType)
	assert.Equal(t, message.TypeHandshakeSuccessful, messages[3].MessageType)
	assert.Equal(t, message.TypeDisconnect, messages[4].MessageType)
}

func TestConnectMessages(t *testing.T) {
	logger := log.NewTestLogger(t)

	dir, err := os.MkdirTemp("temp", "testcase")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	srv, client := createTestServer(t, dir, logger)
	srv.Start()

	connection := client.MustConnect()
	session := connection.MustSession()
	session.MustShell()
	if _, err := session.Write([]byte("Check 1, 2...")); err != nil {
		t.Fatal(err)
	}
	session.ReadRemaining()
	if err := session.Close(); err != nil {
		t.Fatal(err)
	}
	if err := connection.Close(); err != nil {
		t.Fatal(err)
	}

	srv.Stop(10 * time.Second)

	checkStoredAuditMessages(t, dir, logger)
}

func createTestServer(t *testing.T, dir string, logger log.Logger) (sshserver.TestServer, sshserver.TestClient) {
	geoipLookup, err := geoip.New(
		config.GeoIPConfig{
			Provider: "dummy",
		},
	)
	assert.NoError(t, err)

	auditLogHandler, err := auditlogintegration.New(
		config.AuditLogConfig{
			Enable:  true,
			Format:  config.AuditLogFormatBinary,
			Storage: config.AuditLogStorageFile,
			File: config.AuditLogFileConfig{
				Directory: dir,
			},
		},
		&backendHandler{},
		geoipLookup,
		logger,
	)
	assert.NoError(t, err)

	srv := sshserver.NewTestServer(t, auditLogHandler, logger, nil)
	user := sshserver.NewTestUser("test")
	user.RandomPassword()

	client := sshserver.NewTestClient(srv.GetListen(), srv.GetHostKey(), user, logger)
	return srv, client
}

func checkStoredAuditMessages(t *testing.T, dir string, logger log.Logger) {
	messages, errors, done := getStoredMessages(t, dir, logger)
	if done {
		return
	}
	assert.Empty(t, errors)
	assert.NotEmpty(t, messages)
	assert.Equal(t, message.TypeConnect, messages[0].MessageType)
	assert.Equal(t, message.TypeAuthPassword, messages[1].MessageType)
	assert.Equal(t, message.TypeAuthPasswordSuccessful, messages[2].MessageType)
	assert.Equal(t, message.TypeHandshakeSuccessful, messages[3].MessageType)
	assert.Equal(t, message.TypeNewChannelSuccessful, messages[4].MessageType)
	assert.Equal(t, message.TypeChannelRequestShell, messages[5].MessageType)
	assert.Equal(t, message.TypeClose, messages[6].MessageType)
	assert.True(t, messages[7].MessageType == message.TypeExit || messages[7].MessageType == message.TypeDisconnect)
	assert.True(t, messages[8].MessageType == message.TypeExit || messages[8].MessageType == message.TypeDisconnect)
}

func getStoredMessages(t *testing.T, dir string, logger log.Logger) ([]message.Message, []error, bool) {
	storage, err := file.NewStorage(
		config.AuditLogFileConfig{
			Directory: dir,
		}, logger,
	)
	assert.NoError(t, err)
	entryChannel, errChannel := storage.List()
	var logReader io.ReadCloser
	select {
	case err := <-errChannel:
		assert.NoError(t, err)
	case entry := <-entryChannel:
		logReader, err = storage.OpenReader(entry.Name)
		assert.NoError(t, err)
	}
	assert.NotNil(t, logReader)
	if logReader == nil {
		return nil, nil, true
	}

	decoder := binary.NewDecoder()
	messageChannel, errorChannel := decoder.Decode(logReader)
	var messages []message.Message
	var errors []error
loop:
	for {
		select {
		case msg, ok := <-messageChannel:
			if !ok {
				break loop
			}
			messages = append(messages, msg)
		case err, ok := <-errorChannel:
			if !ok {
				break loop
			}
			errors = append(errors, err)
		}
	}
	return messages, errors, false
}

type backendHandler struct {
	session sshserver.SessionChannel
}

func (b *backendHandler) OnAuthKeyboardInteractive(
	meta metadata.ConnectionAuthPendingMetadata,
	challenge func(
		instruction string,
		questions sshserver.KeyboardInteractiveQuestions,
	) (answers sshserver.KeyboardInteractiveAnswers, err error),
) (sshserver.AuthResponse, metadata.ConnectionAuthenticatedMetadata, error) {
	answers, err := challenge(
		"Test",
		sshserver.KeyboardInteractiveQuestions{
			sshserver.KeyboardInteractiveQuestion{
				Question:     "Challenge",
				EchoResponse: true,
			},
		},
	)
	if err != nil {
		return sshserver.AuthResponseFailure, meta.AuthFailed(), err
	}
	answerText, err := answers.GetByQuestionText("Challenge")
	if err == nil {
		if answerText != "Response" {
			return sshserver.AuthResponseUnavailable, meta.AuthFailed(), fmt.Errorf("invalid response")
		}
	}
	return sshserver.AuthResponseSuccess, meta.Authenticated(meta.Username), err
}

func (b *backendHandler) OnClose() {
}

func (b *backendHandler) OnUnsupportedChannelRequest(_ uint64, _ string, _ []byte) {}

func (b *backendHandler) OnFailedDecodeChannelRequest(
	_ uint64,
	_ string,
	_ []byte,
	_ error,
) {
}

func (b *backendHandler) OnEnvRequest(_ uint64, _ string, _ string) error {
	return fmt.Errorf("env requests are not supported")
}

func (b *backendHandler) OnPtyRequest(
	_ uint64,
	_ string,
	_ uint32,
	_ uint32,
	_ uint32,
	_ uint32,
	_ []byte,
) error {
	return fmt.Errorf("pty requests are not supported")
}

func (b *backendHandler) OnExecRequest(
	_ uint64,
	_ string,
) error {
	return fmt.Errorf("exec requests are not supported")
}

func (b *backendHandler) OnShell(
	_ uint64,
) error {
	go func() {
		_, _ = io.ReadAll(b.session.Stdin())
		_, _ = b.session.Stdout().Write([]byte("Hello world!"))
		b.session.ExitStatus(0)
	}()
	return nil
}

func (b *backendHandler) OnSubsystem(
	_ uint64,
	_ string,
) error {
	return fmt.Errorf("subsystem requests are not supported")
}

func (b *backendHandler) OnSignal(_ uint64, _ string) error {
	return fmt.Errorf("signals are not supported")
}

func (b *backendHandler) OnWindow(_ uint64, _ uint32, _ uint32, _ uint32, _ uint32) error {
	return fmt.Errorf("window requests are not supported")
}

func (s *backendHandler) OnX11Request(
	requestID uint64,
	singleConnection bool,
	protocol string,
	cookie string,
	screen uint32,
	reverseHandler sshserver.ReverseForward,
) error {
	return fmt.Errorf("Unimplemented")
}

func (b *backendHandler) OnUnsupportedGlobalRequest(_ uint64, _ string, _ []byte) {
}

func (b *backendHandler) OnFailedDecodeGlobalRequest(_ uint64, _ string, _ []byte, _ error) {

}

func (b *backendHandler) OnUnsupportedChannel(_ uint64, _ string, _ []byte) {
}

func (b *backendHandler) OnSessionChannel(_ metadata.ChannelMetadata, _ []byte, session sshserver.SessionChannel) (
	channel sshserver.SessionChannelHandler,
	failureReason sshserver.ChannelRejection,
) {
	b.session = session
	return b, nil
}

func (s *backendHandler) OnTCPForwardChannel(
	channelID uint64,
	hostToConnect string,
	portToConnect uint32,
	originatorHost string,
	originatorPort uint32,
) (channel sshserver.ForwardChannel, failureReason sshserver.ChannelRejection) {
	return nil, sshserver.NewChannelRejection(ssh.Prohibited, message2.ESSHNotImplemented, "Forwarding channel unimplemented", "Forwarding channel unimplemented")
}

func (s *backendHandler) OnRequestTCPReverseForward(
	bindHost string,
	bindPort uint32,
	reverseHandler sshserver.ReverseForward,
) error {
	return fmt.Errorf("Unimplemented")
}

func (s *backendHandler) OnRequestCancelTCPReverseForward(
	bindHost string,
	bindPort uint32,
) error {
	return fmt.Errorf("Unimplemented")
}

func (s *backendHandler) OnDirectStreamLocal(
	channelID uint64,
	path string,
) (channel sshserver.ForwardChannel, failureReason sshserver.ChannelRejection) {
	return nil, sshserver.NewChannelRejection(ssh.Prohibited, message2.ESSHNotImplemented, "Streamlocal Forwarding unimplemented", "Streamlocal Forwarding unimplemented")
}

func (s *backendHandler) OnRequestStreamLocal(
	path string,
	reverseHandler sshserver.ReverseForward,
) error {
	return fmt.Errorf("Unimplemented")
}

func (s *backendHandler) OnRequestCancelStreamLocal(
	path string,
) error {
	return fmt.Errorf("Unimplemented")
}

func (b *backendHandler) OnAuthPassword(meta metadata.ConnectionAuthPendingMetadata, _ []byte) (
	response sshserver.AuthResponse,
	metadata metadata.ConnectionAuthenticatedMetadata,
	reason error,
) {
	if meta.Username == "test" {
		return sshserver.AuthResponseSuccess, meta.Authenticated(meta.Username), nil
	}
	return sshserver.AuthResponseFailure, meta.AuthFailed(), nil
}

func (b *backendHandler) OnAuthPubKey(meta metadata.ConnectionAuthPendingMetadata, _ auth2.PublicKey) (
	response sshserver.AuthResponse,
	metadata metadata.ConnectionAuthenticatedMetadata,
	reason error,
) {
	return sshserver.AuthResponseFailure, meta.AuthFailed(), nil
}

func (b *backendHandler) OnAuthGSSAPI(_ metadata.ConnectionMetadata) auth.GSSAPIServer {
	return nil
}

func (b *backendHandler) OnHandshakeFailed(_ metadata.ConnectionMetadata, _ error) {}

func (b *backendHandler) OnHandshakeSuccess(meta metadata.ConnectionAuthenticatedMetadata) (
	connection sshserver.SSHConnectionHandler,
	metadata metadata.ConnectionAuthenticatedMetadata,
	failureReason error,
) {
	return b, meta, nil
}

func (b *backendHandler) OnDisconnect() {
}

func (b *backendHandler) OnReady() error {
	return nil
}

func (b *backendHandler) OnShutdown(_ context.Context) {

}

func (b *backendHandler) OnNetworkConnection(
	meta metadata.ConnectionMetadata,
) (sshserver.NetworkConnectionHandler, metadata.ConnectionMetadata, error) {
	return b, meta, nil
}
