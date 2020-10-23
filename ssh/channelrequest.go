package ssh

import (
	"context"
	"github.com/containerssh/containerssh/backend"
	"github.com/containerssh/containerssh/log"
	channelRequest "github.com/containerssh/containerssh/ssh/channel/request"
	"github.com/containerssh/containerssh/ssh/channel/request/env"
	"github.com/containerssh/containerssh/ssh/channel/request/exec"
	"github.com/containerssh/containerssh/ssh/channel/request/pty"
	"github.com/containerssh/containerssh/ssh/channel/request/shell"
	"github.com/containerssh/containerssh/ssh/channel/request/signal"
	"github.com/containerssh/containerssh/ssh/channel/request/subsystem"
	"github.com/containerssh/containerssh/ssh/channel/request/window"
	"github.com/containerssh/containerssh/ssh/server"
	"golang.org/x/crypto/ssh"
)

type channelRequestHandler struct {
	typeHandler *channelRequest.Handler
	session     backend.Session
}

func (handler *channelRequestHandler) OnChannelRequest(ctx context.Context, _ *ssh.ServerConn, channel ssh.Channel, requestType string, payload []byte) server.RequestResponse {
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

type defaultChannelRequestHandlerFactory struct {
	logger log.Logger
}

func NewDefaultChannelRequestHandlerFactory(
	logger log.Logger,
) ChannelRequestHandlerFactory {
	return &defaultChannelRequestHandlerFactory{
		logger: logger,
	}
}

func (factory *defaultChannelRequestHandlerFactory) Make(session backend.Session) *channelRequestHandler {
	handler := channelRequest.NewHandler(factory.logger)
	handler.AddTypeHandler("env", env.New(factory.logger))
	handler.AddTypeHandler("pty-req", pty.New(factory.logger))
	handler.AddTypeHandler("shell", shell.New(factory.logger))
	handler.AddTypeHandler("exec", exec.New(factory.logger))
	handler.AddTypeHandler("subsystem", subsystem.New(factory.logger))
	handler.AddTypeHandler("window-change", window.New(factory.logger))
	handler.AddTypeHandler("signal", signal.New(factory.logger))
	return &channelRequestHandler{
		typeHandler: &handler,
		session:     session,
	}
}
