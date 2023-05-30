package auth

import (
	"context"
	"encoding/base64"
	"time"

	"go.containerssh.io/libcontainerssh/config"
	"go.containerssh.io/libcontainerssh/http"
	"go.containerssh.io/libcontainerssh/log"
	"go.containerssh.io/libcontainerssh/message"
	"go.containerssh.io/libcontainerssh/metadata"
)

func newOIDCProvider(config config.AuthOAuth2ClientConfig, logger log.Logger) (OAuth2Provider, error) {
	return &oidcProvider{
		clientID:     config.ClientID,
		clientSecret: config.ClientSecret,
		config:       config.OIDC,
		logger:       logger,
	}, nil
}

type oidcProvider struct {
	clientID     string
	clientSecret string
	config       config.AuthOIDCConfig
	logger       log.Logger
}

func (o *oidcProvider) SupportsDeviceFlow() bool {
	return o.config.DeviceFlow
}

func (o *oidcProvider) GetDeviceFlow(ctx context.Context, meta metadata.ConnectionAuthPendingMetadata) (OAuth2DeviceFlow, error) {
	flow, err := o.createFlow(ctx, meta)
	if err != nil {
		return nil, err
	}

	return &oidcDeviceFlow{
		flow,
		10 * time.Second,
		"",
		meta,
	}, nil
}

func (o *oidcProvider) SupportsAuthorizationCodeFlow() bool {
	return o.config.AuthorizationCodeFlow
}

func (o *oidcProvider) GetAuthorizationCodeFlow(ctx context.Context, meta metadata.ConnectionAuthPendingMetadata) (OAuth2AuthorizationCodeFlow, error) {
	flow, err := o.createFlow(ctx, meta)
	if err != nil {
		return nil, err
	}

	return &oidcAuthorizationCodeFlow{
		flow,
		meta,
	}, nil
}

func (o *oidcProvider) createFlow(ctx context.Context, meta metadata.ConnectionAuthPendingMetadata) (oidcFlow, error) {
	logger := o.logger.
		WithLabel("connectionID", meta.ConnectionID).
		WithLabel("username", meta.Username)

	cfg := o.config.HTTPClientConfiguration
	cfg.RequestEncoding = config.RequestEncodingWWWURLEncoded
	urlEncodedClient, err := http.NewClientWithHeaders(
		cfg,
		logger,
		map[string][]string{
			"authorization": {
				"Basic " + base64.StdEncoding.EncodeToString([]byte(o.clientID+":"+o.clientSecret)),
			},
		},
		true,
	)
	if err != nil {
		return oidcFlow{}, message.WrapUser(
			err,
			message.EAuthOIDCHTTPClientCreateFailed,
			"Authentication currently unavailable.",
			"Cannot create OIDC device flow authenticator because the HTTP urlEncodedClient configuration failed.",
		)
	}

	discovery := newOIDCDiscovery(o.logger)
	discoveryResponse, err := discovery.Discover(ctx, urlEncodedClient)
	if err != nil {
		return oidcFlow{}, err
	}

	flow := oidcFlow{
		meta:              meta,
		provider:          o,
		connectionID:      meta.ConnectionID,
		username:          meta.Username,
		logger:            logger,
		urlEncodedClient:  urlEncodedClient,
		discoveryResponse: discoveryResponse,
	}
	return flow, nil
}
