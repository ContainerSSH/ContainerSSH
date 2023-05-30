package auth

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"time"

	"go.containerssh.io/libcontainerssh/config"
	http2 "go.containerssh.io/libcontainerssh/http"
	"go.containerssh.io/libcontainerssh/log"
	"go.containerssh.io/libcontainerssh/message"
	"go.containerssh.io/libcontainerssh/metadata"
)

func newGenericProvider(config config.AuthOAuth2ClientConfig, logger log.Logger) (OAuth2Provider, error) {
	return &genericProvider{
		clientID:     config.ClientID,
		clientSecret: config.ClientSecret,
		config:       config.Generic,
		logger:       logger,
	}, nil
}

type genericProvider struct {
	clientID     string
	clientSecret string
	config       config.AuthGenericConfig
	logger       log.Logger
}

func (g genericProvider) SupportsDeviceFlow() bool {
	return false
}

func (g genericProvider) GetDeviceFlow(ctx context.Context, connectionMetadata metadata.ConnectionAuthPendingMetadata) (OAuth2DeviceFlow, error) {
	return nil, fmt.Errorf("the generic provider does not support the device flow")
}

func (g genericProvider) SupportsAuthorizationCodeFlow() bool {
	return true
}

func (g genericProvider) GetAuthorizationCodeFlow(ctx context.Context, connectionMetadata metadata.ConnectionAuthPendingMetadata) (OAuth2AuthorizationCodeFlow, error) {
	flow, err := g.createFlow(ctx, connectionMetadata)
	if err != nil {
		return nil, err
	}

	return &genericAuthorizationCodeFlow{
		flow,
		connectionMetadata,
	}, nil
}

func (g *genericProvider) createFlow(ctx context.Context, meta metadata.ConnectionAuthPendingMetadata) (genericFlow, error) {
	logger := g.logger.
		WithLabel("connectionID", meta.ConnectionID).
		WithLabel("username", meta.Username)

	cfg := g.config.TokenEndpoint
	cfg.RequestEncoding = config.RequestEncodingWWWURLEncoded
	urlEncodedClient, err := http2.NewClientWithHeaders(
		cfg,
		logger,
		map[string][]string{
			"authorization": {
				"Basic " + base64.StdEncoding.EncodeToString([]byte(g.clientID+":"+g.clientSecret)),
			},
		},
		true,
	)
	if err != nil {
		return genericFlow{}, message.WrapUser(
			err,
			message.EAuthOAuth2HTTPClientCreateFailed,
			"Authentication currently unavailable.",
			"Cannot create authenticator because the token endpoint configuration failed.",
		)
	}

	flow := genericFlow{
		meta:         meta,
		client:       urlEncodedClient,
		provider:     g,
		connectionID: meta.ConnectionID,
		username:     meta.Username,
		logger:       logger,
	}
	return flow, nil
}

type genericFlow struct {
	provider     *genericProvider
	client       http2.Client
	connectionID string
	username     string
	logger       log.Logger
	meta         metadata.ConnectionAuthPendingMetadata
	accessToken  string
}

type genericAuthorizationCodeFlow struct {
	genericFlow
	meta metadata.ConnectionAuthPendingMetadata
}

func (g *genericAuthorizationCodeFlow) Deauthorize(ctx context.Context) {
	// oAuth2 doesn't support deauthorization.
}

func (g *genericAuthorizationCodeFlow) GetAuthorizationURL(ctx context.Context) (string, error) {
	endpoint := g.provider.config.AuthorizeEndpointURL
	l, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	query := l.Query()
	query.Set("response_type", "code")
	query.Set("client_id", g.provider.clientID)
	query.Set("scope", "")
	query.Set("state", g.connectionID)
	query.Set("redirect_uri", g.provider.config.RedirectURI)
	l.RawQuery = query.Encode()
	return l.String(), nil
}

type genericAccessTokenRequest struct {
	GrantType   string `json:"grant_type,omitempty" schema:"grant_type"`
	Code        string `json:"code,omitempty" schema:"code"`
	ClientID    string `json:"client_id" schema:"client_id,required"`
	RedirectURI string `json:"redirect_uri" schema:"redirect_uri"`
}

type genericAccessTokenResponse struct {
	AccessToken      string `json:"access_token,omitempty"`
	TokenType        string `json:"token_type,omitempty"`
	ExpiresIn        int    `json:"expires_in,omitempty"`
	RefreshToken     string `json:"refresh_token,omitempty"`
	Scope            string `json:"scope,omitempty"`
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
	ErrorURI         string `json:"error_uri,omitempty"`
}

func (g *genericAuthorizationCodeFlow) Verify(
	ctx context.Context,
	state string,
	authorizationCode string,
) (string, metadata.ConnectionAuthenticatedMetadata, error) {
	accessToken, err := g.getAccessToken(ctx, authorizationCode)
	g.accessToken = accessToken
	if err != nil {
		return "", g.meta.AuthFailed(), err
	}
	m := g.meta.GetMetadata()
	m["OAUTH2_TOKEN"] = metadata.Value{Value: accessToken, Sensitive: true}
	return g.accessToken, g.meta.Authenticated(""), nil
}

func (g *genericAuthorizationCodeFlow) getAccessToken(ctx context.Context, authorizationCode string) (string, error) {
	req := &genericAccessTokenRequest{
		GrantType:   "authorization_code",
		Code:        authorizationCode,
		ClientID:    g.provider.clientID,
		RedirectURI: g.provider.config.RedirectURI,
	}
	resp := &genericAccessTokenResponse{}
	var err error
	var statusCode int
loop:
	for {
		statusCode, err = g.client.Post("", req, resp)
		if err == nil && statusCode == 200 {
			if resp.Error != "" {
				err = message.UserMessage(
					message.EAuthOAuth2AccessTokenFetchFailed,
					"Authentication failed",
					"Error returned from generic oAuth2 server (%s: %s)",
					resp.Error,
					resp.ErrorDescription,
				)
			} else {
				return resp.AccessToken, nil
			}
		}
		err = message.WrapUser(
			err,
			message.EAuthOAuth2AccessTokenFetchFailed,
			"Cannot authenticate at this time.",
			"Non-200 status code from oAuth2 access token API (%d; %s; %s).",
			statusCode,
			resp.Error,
			resp.ErrorDescription,
		)
		if statusCode > 399 && statusCode < 500 {
			return "", err
		}

		g.logger.Debug(err)
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(10 * time.Second):
		}
	}
	err = message.Wrap(
		err,
		message.EAuthGenericTimeout,
		"Cannot authenticate at this time.",
		"Timeout during generic oAuth2 authentication flow.",
	)
	g.logger.Debug(err)
	return "", err
}
