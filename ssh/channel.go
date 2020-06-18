package ssh

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/janoszen/containerssh/backend"
	"github.com/janoszen/containerssh/config"
	configurationClient "github.com/janoszen/containerssh/config/client"
	"github.com/janoszen/containerssh/config/util"
	"github.com/janoszen/containerssh/protocol"
	"github.com/janoszen/containerssh/ssh/server"
	"github.com/qdm12/reprint"
	log "github.com/sirupsen/logrus"
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
}

func NewChannelHandler(
	appConfig *config.AppConfig,
	backendRegistry *backend.Registry,
	configClient configurationClient.ConfigClient,
	channelRequestHandlerFactory ChannelRequestHandlerFactory,
) *ChannelHandler {
	return &ChannelHandler{
		appConfig:                    appConfig,
		backendRegistry:              backendRegistry,
		configClient:                 configClient,
		channelRequestHandlerFactory: channelRequestHandlerFactory,
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
		log.Warnf("Failed to copy application config (%s)", err)
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
			log.Tracef("%v", err)
			return nil, &server.ChannelRejection{
				RejectionReason:  ssh.ResourceShortage,
				RejectionMessage: fmt.Sprintf("internal error while calling the config server: %s", err),
			}
		}
		newConfig, err := util.Merge(&configResponse.Config, &actualConfig)
		if err != nil {
			log.Tracef("%v", err)
			return nil, &server.ChannelRejection{
				RejectionReason:  ssh.ResourceShortage,
				RejectionMessage: fmt.Sprintf("failed to merge config"),
			}
		}
		actualConfig = *newConfig
	}

	containerBackend, err := handler.backendRegistry.GetBackend(actualConfig.Backend)
	if err != nil {
		log.Tracef("%v", err)
		return nil, &server.ChannelRejection{
			RejectionReason:  ssh.ResourceShortage,
			RejectionMessage: fmt.Sprintf("no such backend"),
		}
	}

	backendSession, err := containerBackend.CreateSession(
		string(connection.SessionID()),
		connection.User(),
		&actualConfig,
	)
	if err != nil {
		log.Tracef("%v", err)
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
}

func (factory *channelHandlerFactory) Make(appConfig *config.AppConfig) *ChannelHandler {
	return NewChannelHandler(
		appConfig,
		factory.backendRegistry,
		factory.configClient,
		factory.channelRequestHandlerFactory,
	)
}

func NewDefaultChannelHandlerFactory(
	backendRegistry *backend.Registry,
	configClient configurationClient.ConfigClient,
	channelRequestHandlerFactory ChannelRequestHandlerFactory,
) ChannelHandlerFactory {
	return &channelHandlerFactory{
		backendRegistry:              backendRegistry,
		configClient:                 configClient,
		channelRequestHandlerFactory: channelRequestHandlerFactory,
	}
}
