package server

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"github.com/janoszen/containerssh/log"
	"github.com/janoszen/containerssh/log/writer"
	"golang.org/x/crypto/ssh"
	"net"
	"testing"
)

func createHostKey() (ssh.Signer, error) {
	reader := rand.Reader
	bitSize := 2048
	key, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		return nil, err
	}

	private, err := ssh.NewSignerFromKey(key)
	if err != nil {
		return nil, err
	}

	return private, nil
}

func createTestLogger(
) log.Logger {
	logConfig, err := log.NewConfig(log.StoredConfig{
		Level: "debug",
	})
	if err != nil {
		panic(err)
	}
	logWriter := writer.NewJsonLogWriter()
	return log.NewLoggerPipeline(logConfig, logWriter)
}

func createTestServer(
	readyHandler ReadyHandler,
) (*Server, error) {
	hostKey, err := createHostKey()
	if err != nil {
		return nil, err
	}
	serverConfig := &Config{
		NoClientAuth: true,
		HostKeys: []ssh.Signer{
			hostKey,
		},
	}

	return New(
		"127.0.0.1:0",
		serverConfig,
		readyHandler,
		nil,
		createTestLogger(),
	)
}

type TestHandler struct {
	onReady          func(listener net.Listener)
	onGlobalRequest  func(ctx context.Context, sshConn *ssh.ServerConn, requestType string, payload []byte) RequestResponse
	onChannel        func(ctx context.Context, connection ssh.ConnMetadata, channelType string, extraData []byte) (rejectionReason ssh.RejectionReason, rejectionMessage string)
	onChannelRequest func(ctx context.Context, sshConn *ssh.ServerConn, channel ssh.Channel, requestType string, payload []byte) RequestResponse
}

func (handler *TestHandler) OnReady(listener net.Listener) {
	if handler.onReady == nil {
		return
	}
	handler.onReady(listener)
}

func (handler *TestHandler) OnGlobalRequest(ctx context.Context, sshConn *ssh.ServerConn, requestType string, payload []byte) RequestResponse {
	if handler.onGlobalRequest == nil {
		return RequestResponse{
			Success: false,
			Payload: nil,
		}
	}
	return handler.onGlobalRequest(ctx, sshConn, requestType, payload)
}

func (handler *TestHandler) OnChannel(ctx context.Context, connection ssh.ConnMetadata, channelType string, extraData []byte) (rejectionReason ssh.RejectionReason, rejectionMessage string) {
	if handler.onChannel == nil {
		return ssh.UnknownChannelType, "No channel handler defined"
	}
	return handler.onChannel(ctx, connection, channelType, extraData)
}

func (handler *TestHandler) OnChannelRequest(ctx context.Context, sshConn *ssh.ServerConn, channel ssh.Channel, requestType string, payload []byte) RequestResponse {
	if handler.onChannelRequest == nil {
		return RequestResponse{
			Success: false,
			Payload: nil,
		}
	}
	return handler.onChannelRequest(ctx, sshConn, channel, requestType, payload)
}

func TestEstablishingConnection(t *testing.T) {
	clientErrorChannel := make(chan error)
	serverErrorChannel := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	handler := &TestHandler{
		onReady: func(listener net.Listener) {
			address := listener.Addr().String()
			go func() {
				clientConfig := &ssh.ClientConfig{
					User: "test",
					HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
						return nil
					},
				}
				conn, err := ssh.Dial("tcp", address, clientConfig)
				if err != nil {
					clientErrorChannel <- err
					return
				}
				err = conn.Close()
				if err != nil {
					clientErrorChannel <- err
					return
				}
				// Stop server
				cancel()
				clientErrorChannel <- nil
			}()
		},
	}
	server, err := createTestServer(handler)
	if err != nil {
		t.Fatalf("failed to create test server (%v)", err)
	}

	go func() {
		serverErrorChannel <- server.Run(ctx)
	}()

	for i := 0; i < 2; i++ {
		select {
		case clientError := <-clientErrorChannel:
			if clientError != nil {
				t.Fatalf("an error happened in client (%v)", clientError)
			}
		case serverError := <-serverErrorChannel:
			if serverError != nil {
				t.Fatalf("an error happened in server (%v)", serverError)
			}
		}
	}
}
