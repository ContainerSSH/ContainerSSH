package sshserver

import (
	ssh2 "github.com/containerssh/libcontainerssh/internal/ssh"
	"github.com/containerssh/libcontainerssh/log"
	"golang.org/x/crypto/ssh"
)

type ReverseForwardHandler struct {
	sshConn *ssh.ServerConn
	server  *serverImpl
	logger  log.Logger
}

func (r *ReverseForwardHandler) NewChannelTCP(
	connectedAddress string,
	connectedPort uint32,
	originatorAddress string,
	originatorPort uint32,
) (ForwardChannel, uint64, error) {
	payload := ssh2.ForwardTCPChannelOpenPayload{
		ConnectedAddress:  connectedAddress,
		ConnectedPort:     connectedPort,
		OriginatorAddress: originatorAddress,
		OriginatorPort:    originatorPort,
	}
	mar := ssh.Marshal(payload)
	return r.openChannel(ChannelTypeReverseForward, mar)
}

func (r *ReverseForwardHandler) NewChannelUnix(
	path string,
) (ForwardChannel, uint64, error) {
	payload := ssh2.ForwardedStreamLocalChannelOpenPayload{
		SocketPath: path,
	}
	mar := ssh.Marshal(payload)
	return r.openChannel(ChannelTypeForwardedStreamLocal, mar)
}

func (r *ReverseForwardHandler) NewChannelX11(
	originatorAddress string,
	originatorPort uint32,
) (ForwardChannel, uint64, error) {
	payload := ssh2.X11ChanOpenRequestPayload{
		OriginatorAddress: originatorAddress,
		OriginatorPort:    originatorPort,
	}
	mar := ssh.Marshal(payload)
	return r.openChannel(ChannelTypeX11, mar)
}

func (r *ReverseForwardHandler) openChannel(
	channelType string,
	payload []byte,
) (ForwardChannel, uint64, error) {
	r.server.lock.Lock()
	channelId := r.server.nextChannelID
	r.server.nextChannelID++
	r.server.lock.Unlock()

	sshChan, reqChan, err := r.sshConn.OpenChannel(channelType, payload)
	if err != nil {
		r.logger.Warning("Failed to open forwarding channel", err)
		return nil, 0, err
	}

	go serveRequestChannel(reqChan)

	return sshChan, channelId, nil
}

func serveRequestChannel(c <-chan *ssh.Request) {
	for {
		req, ok := <-c
		if !ok {
			return
		}
		if req.WantReply {
			_ = req.Reply(false, nil)
		}
	}
}
