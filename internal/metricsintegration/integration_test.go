package metricsintegration_test

import (
	"context"
	"fmt"
	"net"
	"testing"

	publicAuth "github.com/containerssh/libcontainerssh/auth"
	"github.com/containerssh/libcontainerssh/config"
	"github.com/containerssh/libcontainerssh/internal/auth"
	"github.com/containerssh/libcontainerssh/internal/geoip/dummy"
	"github.com/containerssh/libcontainerssh/internal/metrics"
	"github.com/containerssh/libcontainerssh/internal/metricsintegration"
	"github.com/containerssh/libcontainerssh/internal/sshserver"
	"github.com/containerssh/libcontainerssh/metadata"
	"github.com/stretchr/testify/assert"
	"github.com/containerssh/libcontainerssh/message"

	"golang.org/x/crypto/ssh"
)

func TestMetricsReporting(t *testing.T) {
	metricsCollector := metrics.New(dummy.New())
	backend := &dummyBackendHandler{
		authResponse: sshserver.AuthResponseSuccess,
	}
	handler, err := metricsintegration.NewHandler(
		config.MetricsConfig{
			Enable: true,
		},
		metricsCollector,
		backend,
	)
	if !assert.NoError(t, err) {
		return
	}
	t.Run("auth=successful", func(t *testing.T) {
		testAuthSuccessful(t, handler, metricsCollector)
	})

	t.Run("auth=failed", func(t *testing.T) {
		testAuthFailed(t, backend, handler, metricsCollector)
	})
}

func testAuthSuccessful(
	t *testing.T,
	handler sshserver.Handler,
	metricsCollector metrics.Collector,
) {
	connectionMeta := metadata.ConnectionMetadata{
		RemoteAddress: metadata.RemoteAddress(
			net.TCPAddr{
				IP:   net.ParseIP("127.0.0.1"),
				Port: 2222,
			},
		),
		ConnectionID: sshserver.GenerateConnectionID(),
		Metadata:     map[string]metadata.Value{},
		Environment:  map[string]metadata.Value{},
		Files:        map[string]metadata.BinaryValue{},
	}

	networkHandler, connectionMeta, err := handler.OnNetworkConnection(connectionMeta)
	if !assert.NoError(t, err) {
		return
	}
	defer networkHandler.OnDisconnect()

	authResponse, meta, err := networkHandler.OnAuthPassword(
		connectionMeta.StartAuthentication("", "foo"),
		[]byte("bar"),
	)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.Equal(t, authResponse, sshserver.AuthResponseSuccess) {
		return
	}
	_, _, err = networkHandler.OnHandshakeSuccess(meta)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, float64(1), metricsCollector.GetMetric(metricsintegration.MetricNameConnections)[0].Value)
	assert.Equal(t, float64(1), metricsCollector.GetMetric(metricsintegration.MetricNameSuccessfulHandshake)[0].Value)
	assert.Equal(t, 0, len(metricsCollector.GetMetric(metricsintegration.MetricNameFailedHandshake)))
	assert.Equal(t, float64(1), metricsCollector.GetMetric(metricsintegration.MetricNameCurrentConnections)[0].Value)

	networkHandler.OnDisconnect()
	assert.Equal(t, float64(1), metricsCollector.GetMetric(metricsintegration.MetricNameConnections)[0].Value)
	assert.Equal(t, float64(0), metricsCollector.GetMetric(metricsintegration.MetricNameCurrentConnections)[0].Value)
}

func testAuthFailed(
	t *testing.T,
	backend *dummyBackendHandler,
	handler sshserver.Handler,
	metricsCollector metrics.Collector,
) {
	connectionMeta := metadata.ConnectionMetadata{
		RemoteAddress: metadata.RemoteAddress(
			net.TCPAddr{
				IP:   net.ParseIP("127.0.0.1"),
				Port: 2222,
			},
		),
		ConnectionID: sshserver.GenerateConnectionID(),
		Metadata:     map[string]metadata.Value{},
		Environment:  map[string]metadata.Value{},
		Files:        map[string]metadata.BinaryValue{},
	}

	assert.Equal(t, float64(1), metricsCollector.GetMetric(metricsintegration.MetricNameConnections)[0].Value)
	assert.Equal(t, float64(0), metricsCollector.GetMetric(metricsintegration.MetricNameCurrentConnections)[0].Value)

	backend.authResponse = sshserver.AuthResponseFailure
	networkHandler, connectionMeta, err := handler.OnNetworkConnection(
		connectionMeta,
	)
	assert.NoError(t, err)
	response, _, err := networkHandler.OnAuthPassword(connectionMeta.StartAuthentication("", "foo"), []byte("bar"))
	assert.NoError(t, err)
	assert.Equal(t, sshserver.AuthResponseFailure, response)
	networkHandler.OnHandshakeFailed(connectionMeta, fmt.Errorf("failed authentication"))
	assert.Equal(t, float64(2), metricsCollector.GetMetric(metricsintegration.MetricNameConnections)[0].Value)
	assert.Equal(t, float64(1), metricsCollector.GetMetric(metricsintegration.MetricNameSuccessfulHandshake)[0].Value)
	assert.Equal(t, float64(1), metricsCollector.GetMetric(metricsintegration.MetricNameFailedHandshake)[0].Value)
	assert.Equal(t, float64(1), metricsCollector.GetMetric(metricsintegration.MetricNameCurrentConnections)[0].Value)

	networkHandler.OnDisconnect()
	assert.Equal(t, float64(2), metricsCollector.GetMetric(metricsintegration.MetricNameConnections)[0].Value)
	assert.Equal(t, float64(0), metricsCollector.GetMetric(metricsintegration.MetricNameCurrentConnections)[0].Value)
}

type dummyBackendHandler struct {
	authResponse sshserver.AuthResponse
}

func (d *dummyBackendHandler) OnClose() {
}

func (d *dummyBackendHandler) OnReady() error {
	return nil
}

func (d *dummyBackendHandler) OnShutdown(_ context.Context) {

}

func (d *dummyBackendHandler) OnNetworkConnection(
	meta metadata.ConnectionMetadata,
) (sshserver.NetworkConnectionHandler, metadata.ConnectionMetadata, error) {
	return d, meta, nil
}

func (d *dummyBackendHandler) OnDisconnect() {
}

func (d *dummyBackendHandler) OnAuthPassword(meta metadata.ConnectionAuthPendingMetadata, _ []byte) (
	response sshserver.AuthResponse,
	metadata metadata.ConnectionAuthenticatedMetadata,
	reason error,
) {
	switch d.authResponse {
	case sshserver.AuthResponseSuccess:
		return d.authResponse, meta.Authenticated(meta.Username), nil
	default:
		return d.authResponse, meta.AuthFailed(), nil
	}
}

func (d *dummyBackendHandler) OnAuthPubKey(meta metadata.ConnectionAuthPendingMetadata, _ publicAuth.PublicKey) (
	response sshserver.AuthResponse,
	metadata metadata.ConnectionAuthenticatedMetadata,
	reason error,
) {
	switch d.authResponse {
	case sshserver.AuthResponseSuccess:
		return d.authResponse, meta.Authenticated(meta.Username), nil
	default:
		return d.authResponse, meta.AuthFailed(), nil
	}
}

func (d *dummyBackendHandler) OnAuthKeyboardInteractive(
	meta metadata.ConnectionAuthPendingMetadata,
	_ func(
		instruction string,
		questions sshserver.KeyboardInteractiveQuestions,
	) (answers sshserver.KeyboardInteractiveAnswers, err error),
) (sshserver.AuthResponse, metadata.ConnectionAuthenticatedMetadata, error) {
	switch d.authResponse {
	case sshserver.AuthResponseSuccess:
		return d.authResponse, meta.Authenticated(meta.Username), nil
	default:
		return d.authResponse, meta.AuthFailed(), nil
	}
}

func (d *dummyBackendHandler) OnAuthGSSAPI(_ metadata.ConnectionMetadata) auth.GSSAPIServer {
	return nil
}

func (d *dummyBackendHandler) OnHandshakeFailed(_ metadata.ConnectionMetadata, _ error) {

}

func (d *dummyBackendHandler) OnHandshakeSuccess(authenticatedMetadata metadata.ConnectionAuthenticatedMetadata) (
	connection sshserver.SSHConnectionHandler,
	metadata metadata.ConnectionAuthenticatedMetadata,
	failureReason error,
) {
	return d, authenticatedMetadata, nil
}

func (d *dummyBackendHandler) OnUnsupportedGlobalRequest(_ uint64, _ string, _ []byte) {

}

func (b *dummyBackendHandler) OnFailedDecodeGlobalRequest(_ uint64, _ string, _ []byte, _ error) {}

func (d *dummyBackendHandler) OnUnsupportedChannel(_ uint64, _ string, _ []byte) {

}

func (d *dummyBackendHandler) OnSessionChannel(
	_ metadata.ChannelMetadata,
	_ []byte,
	session sshserver.SessionChannel,
) (channel sshserver.SessionChannelHandler, failureReason sshserver.ChannelRejection) {
	return &dummySession{
		session: session,
	}, nil
}

func (s *dummyBackendHandler) OnTCPForwardChannel(
	channelID uint64,
	hostToConnect string,
	portToConnect uint32,
	originatorHost string,
	originatorPort uint32,
) (channel sshserver.ForwardChannel, failureReason sshserver.ChannelRejection) {
	return nil, sshserver.NewChannelRejection(ssh.Prohibited, message.ESSHNotImplemented, "Forwading channel unimplemented in docker backend", "Forwading channel unimplemented in docker backend")
}

func (s *dummyBackendHandler) OnRequestTCPReverseForward(
	bindHost string,
	bindPort uint32,
	reverseHandler sshserver.ReverseForward,
) error {
	return fmt.Errorf("Unimplemented")
}

func (s *dummyBackendHandler) OnRequestCancelTCPReverseForward(
	bindHost string,
	bindPort uint32,
) error {
	return fmt.Errorf("Unimplemented")
}

func (s *dummyBackendHandler) OnDirectStreamLocal(
	channelID uint64,
	path string,
) (channel sshserver.ForwardChannel, failureReason sshserver.ChannelRejection) {
	return nil, sshserver.NewChannelRejection(ssh.Prohibited, message.ESSHNotImplemented, "Forwading channel unimplemented in docker backend", "Forwading channel unimplemented in docker backend")
}

func (s *dummyBackendHandler) OnRequestStreamLocal(
	path string,
	reverseHandler sshserver.ReverseForward,
) error {
	return fmt.Errorf("Unimplemented")
}

func (s *dummyBackendHandler) OnRequestCancelStreamLocal(
	path string,
) error {
	return fmt.Errorf("Unimplemented")
}

type dummySession struct {
	session sshserver.SessionChannel
}

func (d *dummySession) OnClose() {
}

func (d *dummySession) OnShutdown(_ context.Context) {
}

func (d *dummySession) OnUnsupportedChannelRequest(_ uint64, _ string, _ []byte) {

}

func (d *dummySession) OnFailedDecodeChannelRequest(
	_ uint64,
	_ string,
	_ []byte,
	_ error,
) {

}

func (d *dummySession) OnEnvRequest(_ uint64, _ string, _ string) error {
	return fmt.Errorf("env not supported")
}

func (d *dummySession) OnPtyRequest(
	_ uint64,
	_ string,
	_ uint32,
	_ uint32,
	_ uint32,
	_ uint32,
	_ []byte,
) error {
	return fmt.Errorf("PTY not supported")
}

func (d *dummySession) OnExecRequest(
	_ uint64,
	exec string,
) error {
	go func() {
		_, err := d.session.Stdout().Write([]byte(fmt.Sprintf("Exec request received: %s", exec)))
		if err != nil {
			d.session.ExitStatus(2)
		} else {
			d.session.ExitStatus(0)
		}
	}()
	return nil
}

func (d *dummySession) OnShell(
	_ uint64,
) error {
	return fmt.Errorf("shell not supported")
}

func (d *dummySession) OnSubsystem(
	_ uint64,
	subsystem string,
) error {
	if subsystem != "sftp" {
		return fmt.Errorf("subsystem not supported")
	}
	go func() {
		_, err := d.session.Stdout().Write([]byte(fmt.Sprintf("Subsystem request received: %s", subsystem)))
		if err != nil {
			d.session.ExitStatus(2)
		} else {
			d.session.ExitStatus(0)
		}
	}()
	return nil
}

func (d *dummySession) OnSignal(_ uint64, _ string) error {
	return fmt.Errorf("signal not supported")
}

func (d *dummySession) OnWindow(
	_ uint64,
	_ uint32,
	_ uint32,
	_ uint32,
	_ uint32,
) error {
	return fmt.Errorf("window changes are not supported")
}

func (s *dummySession) OnX11Request(
	requestID uint64,
	singleConnection bool,
	protocol string,
	cookie string,
	screen uint32,
	reverseHandler sshserver.ReverseForward,
) error {
	return fmt.Errorf("Unimplemented")
}
