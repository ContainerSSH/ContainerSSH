package ssh

import (
	"github.com/containerssh/containerssh/config"
	"github.com/containerssh/containerssh/ssh/server"
	"golang.org/x/crypto/ssh"
)

type ConnectionHandler struct {
	config                      *config.AppConfig
	globalRequestHandlerFactory GlobalRequestHandlerFactory
	channelHandlerFactory       ChannelHandlerFactory
}

func NewConnectionHandler(
	config *config.AppConfig,
	globalRequestHandlerFactory GlobalRequestHandlerFactory,
	channelHandlerFactory ChannelHandlerFactory,
) *ConnectionHandler {
	return &ConnectionHandler{
		config:                      config,
		globalRequestHandlerFactory: globalRequestHandlerFactory,
		channelHandlerFactory:       channelHandlerFactory,
	}
}

func (handler ConnectionHandler) OnConnection(
	_ *ssh.ServerConn,
) (server.GlobalRequestHandler, server.ChannelHandler, error) {
	return handler.globalRequestHandlerFactory.Make(), handler.channelHandlerFactory.Make(handler.config), nil
}
