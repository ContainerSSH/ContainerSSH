package sshserver_test

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/pires/go-proxyproto"
	"github.com/stretchr/testify/assert"
	"go.containerssh.io/containerssh/config"
	"go.containerssh.io/containerssh/internal/sshserver"
	"go.containerssh.io/containerssh/internal/structutils"
	"go.containerssh.io/containerssh/log"
	"go.containerssh.io/containerssh/metadata"
)

//region Tests

func TestProxyProtocolHeader(t *testing.T) {
	server, handler := newProxyProtocolServer(t, []string{"127.0.0.0/8"})

	assert.Equal(t, "1.2.3.4", sendProxyProtocolConnection(t, server, handler, newProxyProtocolHeader(t, "1.2.3.4")))
}

func TestProxyProtocolNoHeader(t *testing.T) {
	server, handler := newProxyProtocolServer(t, []string{"127.0.0.0/8"})

	assert.Equal(t, "127.0.0.1", sendProxyProtocolConnection(t, server, handler, []byte("SSH-2.0-Test\r\n")))
}

func TestProxyProtocolUntrustedHeader(t *testing.T) {
	server, handler := newProxyProtocolServer(t, []string{"10.0.0.0/8"})

	assert.Equal(t, "127.0.0.1", sendProxyProtocolConnection(t, server, handler, newProxyProtocolHeader(t, "1.2.3.4")))
}

//endregion

//region Helper

func newProxyProtocolServer(t *testing.T, allowedCIDRs []string) (sshserver.TestServer, *recordHandler) {
	t.Helper()

	cfg := &config.SSHConfig{}
	structutils.Defaults(cfg)
	cfg.ProxyProtocolAllowedCIDRs = allowedCIDRs

	handler := &recordHandler{connected: make(chan metadata.ConnectionMetadata, 1)}
	server := sshserver.NewTestServer(t, handler, log.NewTestLogger(t), cfg)
	server.Start()
	return server, handler
}

func newProxyProtocolHeader(t *testing.T, sourceIP string) []byte {
	t.Helper()

	header, err := (&proxyproto.Header{
		Version:           2,
		Command:           proxyproto.PROXY,
		TransportProtocol: proxyproto.TCPv4,
		SourceAddr:        &net.TCPAddr{IP: net.ParseIP(sourceIP), Port: 1234},
		DestinationAddr:   &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 2222},
	}).Format()
	if err != nil {
		t.Fatalf("failed to format the PROXY protocol header (%v)", err)
	}
	return header
}

func sendProxyProtocolConnection(
	t *testing.T, server sshserver.TestServer, handler *recordHandler, data []byte,
) string {
	t.Helper()

	conn, err := net.Dial("tcp", server.GetListen())
	if err != nil {
		t.Fatalf("failed to connect to the server (%v)", err)
	}
	defer func() { _ = conn.Close() }()
	if _, err := conn.Write(data); err != nil {
		t.Fatalf("failed to write to the server (%v)", err)
	}

	select {
	case meta := <-handler.connected:
		return meta.RemoteAddress.IP.String()
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for the connection")
		return ""
	}
}

//endregion

//region Handlers

type recordHandler struct {
	sshserver.AbstractHandler

	connected chan metadata.ConnectionMetadata
}

func (r *recordHandler) OnNetworkConnection(meta metadata.ConnectionMetadata) (
	sshserver.NetworkConnectionHandler, metadata.ConnectionMetadata, error,
) {
	r.connected <- meta
	return nil, meta, fmt.Errorf("not accepting connections")
}

//endregion
