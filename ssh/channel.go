package ssh

import (
	"context"
	"fmt"

	"encoding/base64"

	"github.com/janoszen/containerssh/backend"
	"github.com/janoszen/containerssh/config"
	configurationClient "github.com/janoszen/containerssh/config/client"
	"github.com/janoszen/containerssh/config/util"
	"github.com/janoszen/containerssh/log"
	"github.com/janoszen/containerssh/protocol"
	"github.com/janoszen/containerssh/ssh/server"

	"github.com/qdm12/reprint"
	"golang.org/x/crypto/ssh"
)

type ChannelHandlerFactory interface {
	Make(appConfig *config.AppConfig) *ChannelHandler
}

type ChannelHandler struct {
	appConfig                    *config.AppConfig
	backendRegistry              *backend.Registry
	configClient                 configurationClient.ConfigClient
	channelRequestHandlerFactory ChannelRequestHandlerFactory
	logger 						 log.Logger
}

func NewChannelHandler(
	appConfig *config.AppConfig,
	backendRegistry *backend.Registry,
	configClient configurationClient.ConfigClient,
	channelRequestHandlerFactory ChannelRequestHandlerFactory,
	logger 						 log.Logger,
) *ChannelHandler {
	return &ChannelHandler{
		appConfig:                    appConfig,
		backendRegistry:              backendRegistry,
		configClient:                 configClient,
		channelRequestHandlerFactory: channelRequestHandlerFactory,
		logger:                       logger,
	}
}

func (handler *ChannelHandler) OnChannel(
	ctx context.Context,
	connection ssh.ConnMetadata,
	channelType string,
	extraData []byte,
) (server.ChannelRequestHandler, *server.ChannelRejection) {
	if channelType != "session" {
		return nil, &server.ChannelRejection{
			RejectionReason:  ssh.UnknownChannelType,
			RejectionMessage: "unknown channel type",
		}
	}

	actualConfig := config.AppConfig{}
	err := reprint.FromTo(handler.appConfig, &actualConfig)
	if err != nil {
		handler.logger.WarningF("failed to copy application config (%v)", err)
		return nil, &server.ChannelRejection{
			RejectionReason:  ssh.ResourceShortage,
			RejectionMessage: "failed to create config",
		}
	}

	if handler.configClient != nil {
		configResponse, err := handler.configClient.GetConfig(protocol.ConfigRequest{
			Username:  connection.User(),
			SessionId: base64.StdEncoding.EncodeToString(connection.SessionID()),
		})
		if err != nil {
			handler.logger.DebugE(err)
			return nil, &server.ChannelRejection{
				RejectionReason:  ssh.ResourceShortage,
				RejectionMessage: fmt.Sprintf("internal error while calling the config server: %s", err),
			}
		}
		newConfig, err := util.Merge(&configResponse.Config, &actualConfig)
		if err != nil {
			handler.logger.DebugE(err)
			return nil, &server.ChannelRejection{
				RejectionReason:  ssh.ResourceShortage,
				RejectionMessage: fmt.Sprintf("failed to merge config"),
			}
		}
		actualConfig = *newConfig
	}

	containerBackend, err := handler.backendRegistry.GetBackend(actualConfig.Backend)
	if err != nil {
		handler.logger.DebugE(err)
		return nil, &server.ChannelRejection{
			RejectionReason:  ssh.ResourceShortage,
			RejectionMessage: fmt.Sprintf("no such backend"),
		}
	}

	backendSession, err := containerBackend.CreateSession(
		string(connection.SessionID()),
		connection.User(),
		&actualConfig,
		handler.logger,
	)
	if err != nil {
		handler.logger.DebugE(err)
		return nil, &server.ChannelRejection{
			RejectionReason:  ssh.ResourceShortage,
			RejectionMessage: fmt.Sprintf("internal error while creating backend"),
		}
	}

	return handler.channelRequestHandlerFactory.Make(backendSession), nil
}

type channelHandlerFactory struct {
	backendRegistry              *backend.Registry
	configClient                 configurationClient.ConfigClient
	channelRequestHandlerFactory ChannelRequestHandlerFactory
	logger                       log.Logger
	loggerFactory                log.LoggerFactory
}

func (factory *channelHandlerFactory) Make(appConfig *config.AppConfig) *ChannelHandler {
	logConfig, err := log.NewConfig(appConfig.Log.Level)
	logger := factory.logger
	if err != nil {
		factory.logger.WarningF("invalid log level (%s) using default logger", appConfig.Log.Level)
	} else {
		logger = factory.loggerFactory.Make(logConfig)
	}

	return NewChannelHandler(
		appConfig,
		factory.backendRegistry,
		factory.configClient,
		factory.channelRequestHandlerFactory,
		logger,
	)
}

func NewDefaultChannelHandlerFactory(
	backendRegistry *backend.Registry,
	configClient configurationClient.ConfigClient,
	channelRequestHandlerFactory ChannelRequestHandlerFactory,
	logger                       log.Logger,
	loggerFactory                log.LoggerFactory,
) ChannelHandlerFactory {
	return &channelHandlerFactory{
		backendRegistry:              backendRegistry,
		configClient:                 configClient,
		channelRequestHandlerFactory: channelRequestHandlerFactory,
		logger: logger,
		loggerFactory: loggerFactory,
	}
}
