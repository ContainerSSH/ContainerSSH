package agentprotocol_test

import (
	"fmt"
	"io"
	"testing"

	proto "github.com/containerssh/libcontainerssh/agentprotocol"

	log "github.com/containerssh/libcontainerssh/log"
)

// region Tests
func TestConnectionSetup(t *testing.T) {
	log := log.NewTestLogger(t)
	fromClientReader, fromClientWriter := io.Pipe()
	toClientReader, toClientWriter := io.Pipe()

	clientCtx := proto.NewForwardCtx(toClientReader, fromClientWriter, log)

	serverCtx := proto.NewForwardCtx(fromClientReader, toClientWriter, log)

	closeChan := make(chan struct{})

	go func() {
		connChan, err := serverCtx.StartReverseForwardClient(
			"127.0.0.1",
			8080,
			false,
		)
		if err != nil {
			panic(err)
		}

		testConServer := <-connChan
		err = testConServer.Accept()
		if err != nil {
			log.Error("Error accept connection", err)
		}
		buf := make([]byte, 512)
		nBytes, err := testConServer.Read(buf)
		if err != nil {
			log.Error("Failed to read from server")
		}
		_, err = testConServer.Write(buf[:nBytes])
		if err != nil {
			log.Error("Failed to write to server")
		}
		<-closeChan
		serverCtx.Kill()
	}()

	conType, setup, connectionChan, err := clientCtx.StartClient()
	if err != nil {
		t.Fatal("Test failed with error", err)
	}
	if conType != proto.CONNECTION_TYPE_PORT_FORWARD {
		panic(fmt.Errorf("Invalid connection type %d", conType))
	}

	go func() {
		for {
			conn, ok := <-connectionChan
			if !ok {
				break
			}
			_ = conn.Reject()
			_ = conn.Close()
		}
	}()

	testConClient, err := clientCtx.NewConnectionTCP(
		setup.BindHost,
		setup.BindPort,
		"127.0.0.5",
		8081,
		func() error {
			return nil
		},
	)
	if err != nil {
		panic(err)
	}

	buf := make([]byte, 512)
	_, err = testConClient.Write([]byte("Message to server"))
	if err != nil {
		t.Fatal(err)
	}
	nBytes, err := testConClient.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	if string(buf[:nBytes]) != "Message to server" {
		t.Fatalf("Expected to read 'Message to server' instead got %s", string(buf[:nBytes]))
	}
	_ = testConClient.Close()
	close(closeChan)
	clientCtx.Kill()
}
