package auth

import (
	"context"
	"time"

	"github.com/containerssh/containerssh/config"
	http2 "github.com/containerssh/containerssh/http"
	"github.com/containerssh/containerssh/log"
	"github.com/containerssh/containerssh/message"
)

//region Config

// AuthOIDCConfig is the configuration for OpenID Connect authentication.


//endregion

//region GeoIPProvider

func newOIDCProvider(config config.AuthConfig, logger log.Logger) (OAuth2Provider, error) {
	return &oidcProvider{
		config: config,
		logger: logger,
	}, nil
}

type oidcProvider struct {
	config config.AuthConfig
	logger log.Logger
}

func (o *oidcProvider) SupportsDeviceFlow() bool {
	return o.config.OAuth2.OIDC.DeviceFlow
}

func (o *oidcProvider) GetDeviceFlow(connectionID string, username string) (OAuth2DeviceFlow, error) {
	flow, err := o.createFlow(connectionID, username)
	if err != nil {
		return nil, err
	}

	return &oidcDeviceFlow{
		flow,
	}, nil
}

func (o *oidcProvider) SupportsAuthorizationCodeFlow() bool {
	return o.config.OAuth2.OIDC.AuthorizationCodeFlow
}

func (o *oidcProvider) GetAuthorizationCodeFlow(connectionID string, username string) (
	OAuth2AuthorizationCodeFlow,
	error,
) {
	flow, err := o.createFlow(connectionID, username)
	if err != nil {
		return nil, err
	}

	return &oidcAuthorizationCodeFlow{
		flow,
	}, nil
}

func (o *oidcProvider) createFlow(connectionID string, username string) (oidcFlow, error) {
	logger := o.logger.WithLabel("connectionID", connectionID).WithLabel("username", username)

	client, err := http2.NewClient(o.config.OAuth2.OIDC.HTTPClientConfiguration, logger)
	if err != nil {
		return oidcFlow{}, message.WrapUser(
			err,
			EGitHubHTTPClientCreateFailed,
			"Authentication currently unavailable.",
			"Cannot create GitHub device flow authenticator because the HTTP client configuration failed.",
		)
	}

	flow := oidcFlow{
		provider:     o,
		connectionID: connectionID,
		username:     username,
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
	logger log.Logger
	client http2.Client
}

func (o *oidcFlow) Deauthorize(ctx context.Context) {
	panic("implement me")
}

//endregion

//region Device flow

type oidcDeviceFlow struct {
	oidcFlow
}

func (o *oidcDeviceFlow) GetAuthorizationURL(ctx context.Context) (
	verificationLink string,
	userCode string,
	expiration time.Duration,
	err error,
) {
	panic("implement me")
}

func (o *oidcDeviceFlow) Verify(ctx context.Context) (map[string]string, error) {
	panic("implement me")
}

//endregion

//region Authorization code flow

type oidcAuthorizationCodeFlow struct {
	oidcFlow
}

func (o *oidcAuthorizationCodeFlow) GetAuthorizationURL(ctx context.Context) (string, error) {
	panic("implement me")
}

func (o *oidcAuthorizationCodeFlow) Verify(
	ctx context.Context,
	state string,
	authorizationCode string,
) (map[string]string, error) {
	panic("implement me")
}

//endregion