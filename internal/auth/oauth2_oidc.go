package auth

import (
	"context"
	"time"

    "go.containerssh.io/libcontainerssh/config"
    http2 "go.containerssh.io/libcontainerssh/http"
    "go.containerssh.io/libcontainerssh/log"
    "go.containerssh.io/libcontainerssh/message"
    "go.containerssh.io/libcontainerssh/metadata"
)

//region Config

// AuthOIDCConfig is the configuration for OpenID Connect authentication.

//endregion

//region GeoIPProvider

func newOIDCProvider(config config.AuthOIDCConfig, logger log.Logger) (OAuth2Provider, error) {
	return &oidcProvider{
		config: config,
		logger: logger,
	}, nil
}

type oidcProvider struct {
	config config.AuthOIDCConfig
	logger log.Logger
}

func (o *oidcProvider) SupportsDeviceFlow() bool {
	return o.config.DeviceFlow
}

func (o *oidcProvider) GetDeviceFlow(meta metadata.ConnectionAuthPendingMetadata) (OAuth2DeviceFlow, error) {
	flow, err := o.createFlow(meta)
	if err != nil {
		return nil, err
	}

	return &oidcDeviceFlow{
		flow,
		meta,
	}, nil
}

func (o *oidcProvider) SupportsAuthorizationCodeFlow() bool {
	return o.config.AuthorizationCodeFlow
}

func (o *oidcProvider) GetAuthorizationCodeFlow(meta metadata.ConnectionAuthPendingMetadata) (
	OAuth2AuthorizationCodeFlow,
	error,
) {
	flow, err := o.createFlow(meta)
	if err != nil {
		return nil, err
	}

	return &oidcAuthorizationCodeFlow{
		flow,
		meta,
	}, nil
}

func (o *oidcProvider) createFlow(meta metadata.ConnectionAuthPendingMetadata) (oidcFlow, error) {
	logger := o.logger.WithLabel("connectionID", meta.ConnectionID).WithLabel("username", meta.Username)

	client, err := http2.NewClient(o.config.HTTPClientConfiguration, logger)
	if err != nil {
		return oidcFlow{}, message.WrapUser(
			err,
			message.EAuthGitHubHTTPClientCreateFailed,
			"Authentication currently unavailable.",
			"Cannot create GitHub device flow authenticator because the HTTP client configuration failed.",
		)
	}

	flow := oidcFlow{
		meta:         meta,
		provider:     o,
		connectionID: meta.ConnectionID,
		username:     meta.Username,
		logger:       logger,
		client:       client,
	}
	return flow, nil
}

//endregion

//region Flow

type oidcFlow struct {
	provider     *oidcProvider
	connectionID string
	username     string
	logger       log.Logger
	client       http2.Client
	meta         metadata.ConnectionAuthPendingMetadata
}

func (o *oidcFlow) Deauthorize(ctx context.Context) {
	panic("implement me")
}

//endregion

//region Device flow

type oidcDeviceFlow struct {
	oidcFlow
	meta metadata.ConnectionAuthPendingMetadata
}

func (o *oidcDeviceFlow) GetAuthorizationURL(ctx context.Context) (
	verificationLink string,
	userCode string,
	expiration time.Duration,
	err error,
) {
	panic("implement me")
}

func (o *oidcDeviceFlow) Verify(ctx context.Context) (string, metadata.ConnectionAuthenticatedMetadata, error) {
	panic("implement me")
}

//endregion

//region Authorization code flow

type oidcAuthorizationCodeFlow struct {
	oidcFlow
	meta metadata.ConnectionAuthPendingMetadata
}

func (o *oidcAuthorizationCodeFlow) GetAuthorizationURL(ctx context.Context) (string, error) {
	panic("implement me")
}

func (o *oidcAuthorizationCodeFlow) Verify(
	ctx context.Context,
	state string,
	authorizationCode string,
) (string, metadata.ConnectionAuthenticatedMetadata, error) {
	panic("implement me")
}

//endregion
