package ssh

import (
	"context"
	"github.com/janoszen/containerssh/backend"
	channelRequest "github.com/janoszen/containerssh/ssh/channel/request"
	"github.com/janoszen/containerssh/ssh/channel/request/env"
	"github.com/janoszen/containerssh/ssh/channel/request/pty"
	"github.com/janoszen/containerssh/ssh/channel/request/run"
	"github.com/janoszen/containerssh/ssh/channel/request/signal"
	"github.com/janoszen/containerssh/ssh/channel/request/subsystem"
	"github.com/janoszen/containerssh/ssh/channel/request/window"
	"github.com/janoszen/containerssh/ssh/server"
	"golang.org/x/crypto/ssh"
)

type channelRequestHandler struct {
	typeHandler *channelRequest.Handler
	session     backend.Session
}

func (handler *channelRequestHandler) OnChannelRequest(ctx context.Context, sshConn *ssh.ServerConn, channel ssh.Channel, requestType string, payload []byte) server.RequestResponse {
	responseChannel := make(chan server.RequestResponse)
	reply := func(success bool, message []byte) {
		responseChannel <- server.RequestResponse{
			Success: success,
			Payload: message,
		}
	}
	go handler.typeHandler.OnChannelRequest(
		requestType,
		payload,
		reply,
		channel,
		handler.session,
	)
	select {
	case response := <-responseChannel:
		return response
	case <-ctx.Done():
		return server.RequestResponse{
			Success: false,
			Payload: []byte("server is shutting down"),
		}
	}
}

type ChannelRequestHandlerFactory interface {
	Make(session backend.Session) *channelRequestHandler
}

type defaultChannelRequestHandlerFactory struct{}

func NewDefaultChannelRequestHandlerFactory() ChannelRequestHandlerFactory {
	return &defaultChannelRequestHandlerFactory{}
}

func (factory *defaultChannelRequestHandlerFactory) Make(session backend.Session) *channelRequestHandler {
	handler := channelRequest.NewHandler()
	handler.AddTypeHandler("env", env.RequestTypeHandler)
	handler.AddTypeHandler("pty-req", pty.RequestTypeHandler)
	handler.AddTypeHandler("shell", run.ShellRequestTypeHandler)
	handler.AddTypeHandler("exec", run.ExecRequestTypeHandler)
	handler.AddTypeHandler("subsystem", subsystem.RequestTypeHandler)
	handler.AddTypeHandler("window-change", window.RequestTypeHandler)
	handler.AddTypeHandler("signal", signal.RequestTypeHandler)
	return &channelRequestHandler{
		typeHandler: &handler,
		session:     session,
	}
}
