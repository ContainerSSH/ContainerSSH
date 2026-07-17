package sshserver

import (
	"sync"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestRemoveClientConnection(t *testing.T) {
	disconnected := &ssh.ServerConn{}
	connected := &ssh.ServerConn{}
	server := &serverImpl{
		lock: &sync.Mutex{},
		clientSockets: map[*ssh.ServerConn]bool{
			disconnected: true,
			connected:    true,
		},
		connMap: map[string]connection{
			"disconnected": {sshConn: disconnected},
			"connected":    {sshConn: connected},
		},
	}

	server.removeClientConnection("disconnected", disconnected)

	if _, ok := server.clientSockets[disconnected]; ok {
		t.Fatal("disconnected SSH client was not removed from clientSockets")
	}
	if _, ok := server.connMap["disconnected"]; ok {
		t.Fatal("disconnected SSH client was not removed from connMap")
	}
	if _, ok := server.clientSockets[connected]; !ok {
		t.Fatal("connected SSH client was removed from clientSockets")
	}
	if _, ok := server.connMap["connected"]; !ok {
		t.Fatal("connected SSH client was removed from connMap")
	}
}
