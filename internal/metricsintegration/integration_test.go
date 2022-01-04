package metricsintegration_test

import (
	"context"
	"fmt"
	"net"
	"testing"

	auth2 "github.com/containerssh/libcontainerssh/auth"
	"github.com/containerssh/libcontainerssh/config"
	"github.com/containerssh/libcontainerssh/internal/auth"
	"github.com/containerssh/libcontainerssh/internal/geoip/dummy"
	"github.com/containerssh/libcontainerssh/internal/metrics"
	"github.com/containerssh/libcontainerssh/internal/sshserver"
	"github.com/stretchr/testify/assert"

	"github.com/containerssh/libcontainerssh/internal/metricsintegration"
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
	networkHandler, err := handler.OnNetworkConnection(
		net.TCPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 2222,
		},
		sshserver.GenerateConnectionID(),
	)
	if !assert.NoError(t, err) {
		return
	}
	defer networkHandler.OnDisconnect()

	authResponse, metadata, err := networkHandler.OnAuthPassword("foo", []byte("bar"), "")
	if !assert.NoError(t, err) {
		return
	}
	if !assert.Equal(t, authResponse, sshserver.AuthResponseSuccess) {
		return
	}
	_, err = networkHandler.OnHandshakeSuccess("foo", "", metadata)
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
	assert.Equal(t, float64(1), metricsCollector.GetMetric(metricsintegration.MetricNameConnections)[0].Value)
	assert.Equal(t, float64(0), metricsCollector.GetMetric(metricsintegration.MetricNameCurrentConnections)[0].Value)

	backend.authResponse = sshserver.AuthResponseFailure
	networkHandler, err := handler.OnNetworkConnection(
		net.TCPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 2222,
		},
		sshserver.GenerateConnectionID(),
	)
	assert.NoError(t, err)
	response, _, err := networkHandler.OnAuthPassword("foo", []byte("bar"), "")
	assert.NoError(t, err)
	assert.Equal(t, sshserver.AuthResponseFailure, response)
	networkHandler.OnHandshakeFailed(fmt.Errorf("failed authentication"))
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
	_ net.TCPAddr,
	_ string,
) (sshserver.NetworkConnectionHandler, error) {
	return d, nil
}

func (d *dummyBackendHandler) OnDisconnect() {
}

func (d *dummyBackendHandler) OnAuthPassword(_ string, _ []byte, _ string) (
	response sshserver.AuthResponse,
	metadata *auth2.ConnectionMetadata,
	reason error,
) {
	return d.authResponse, nil, nil
}

func (d *dummyBackendHandler) OnAuthPubKey(_ string, _ string, _ string) (
	response sshserver.AuthResponse,
	metadata *auth2.ConnectionMetadata,
	reason error,
) {
	return d.authResponse, nil, nil
}

func (d *dummyBackendHandler) OnAuthKeyboardInteractive(
	_ string,
	_ func(
		instruction string,
		questions sshserver.KeyboardInteractiveQuestions,
	) (answers sshserver.KeyboardInteractiveAnswers, err error),
	_ string,
) (response sshserver.AuthResponse, metadata *auth2.ConnectionMetadata, reason error) {
	return d.authResponse, nil, nil
}

func (d *dummyBackendHandler) OnAuthGSSAPI() auth.GSSAPIServer {
	return nil
}

func (d *dummyBackendHandler) OnHandshakeFailed(_ error) {

}

func (d *dummyBackendHandler) OnHandshakeSuccess(_ string, _ string, _ *auth2.ConnectionMetadata) (
	connection sshserver.SSHConnectionHandler,
	failureReason error,
) {
	return d, nil
}

func (d *dummyBackendHandler) OnUnsupportedGlobalRequest(_ uint64, _ string, _ []byte) {

}

func (d *dummyBackendHandler) OnUnsupportedChannel(_ uint64, _ string, _ []byte) {

}

func (d *dummyBackendHandler) OnSessionChannel(
	_ uint64,
	_ []byte,
	session sshserver.SessionChannel,
) (channel sshserver.SessionChannelHandler, failureReason sshserver.ChannelRejection) {
	return &dummySession{
		session: session,
	}, nil
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
