package auth

import (
	"context"
	"net/url"
	"time"

	"go.containerssh.io/libcontainerssh/message"
	"go.containerssh.io/libcontainerssh/metadata"
)

type oidcAuthorizationCodeFlow struct {
	oidcFlow
	meta metadata.ConnectionAuthPendingMetadata
}

func (o *oidcAuthorizationCodeFlow) GetAuthorizationURL(ctx context.Context) (string, error) {
	endpoint := o.discoveryResponse.AuthorizationEndpoint
	link, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	query := link.Query()
	query.Set("response_type", "code")
	query.Set("client_id", o.provider.clientID)
	query.Set("scope", "openid")
	query.Set("state", o.connectionID)
	query.Set("redirect_uri", o.provider.config.RedirectURI)
	link.RawQuery = query.Encode()
	return link.String(), nil
}

type oidcAccessTokenRequest struct {
	GrantType   string `json:"grant_type,omitempty" schema:"grant_type"`
	Code        string `json:"code,omitempty" schema:"code"`
	ClientID    string `json:"client_id" schema:"client_id,required"`
	DeviceCode  string `json:"device_code" schema:"device_code"`
	RedirectURI string `json:"redirect_uri" schema:"redirect_uri"`
}

type oidcAccessTokenResponse struct {
	AccessToken      string `json:"access_token,omitempty"`
	TokenType        string `json:"token_type,omitempty"`
	ExpiresIn        int    `json:"expires_in,omitempty"`
	RefreshToken     string `json:"refresh_token,omitempty"`
	Scope            string `json:"scope,omitempty"`
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
	ErrorURI         string `json:"error_uri,omitempty"`
}

func (o *oidcAuthorizationCodeFlow) Verify(
	ctx context.Context,
	state string,
	authorizationCode string,
) (string, metadata.ConnectionAuthenticatedMetadata, error) {
	accessToken, err := o.getAccessToken(ctx, authorizationCode)
	o.accessToken = accessToken
	if err != nil {
		if accessToken != "" {
			o.Deauthorize(ctx)
		}
		return "", o.meta.AuthFailed(), err
	}
	return o.getIdentity(ctx, o.meta, accessToken)
}

func (o *oidcAuthorizationCodeFlow) getAccessToken(ctx context.Context, authorizationCode string) (string, error) {
	endpoint := o.discoveryResponse.TokenEndpoint
	req := &oidcAccessTokenRequest{
		GrantType:   "authorization_code",
		Code:        authorizationCode,
		ClientID:    o.provider.clientID,
		RedirectURI: o.provider.config.RedirectURI,
	}
	resp := &oidcAccessTokenResponse{}
	var err error
	var statusCode int
loop:
	for {
		statusCode, err = o.urlEncodedClient.RequestURL("POST", endpoint, req, resp)
		if err == nil && statusCode == 200 {
			if resp.Error != "" {
				err = message.UserMessage(
					message.EAuthOIDCAccessTokenFetchFailed,
					"Authentication failed",
					"Error returned from OIDC server (%s: %s)",
					resp.Error,
					resp.ErrorDescription,
				)
			} else {
				return resp.AccessToken, nil
			}
		}
		err = message.WrapUser(
			err,
			message.EAuthOIDCAccessTokenFetchFailed,
			"Cannot authenticate at this time.",
			"Non-200 status code from OIDC access token API (%d; %s; %s).",
			statusCode,
			resp.Error,
			resp.ErrorDescription,
		)
		if statusCode > 399 && statusCode < 500 {
			return "", err
		}

		o.logger.Debug(err)
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(10 * time.Second):
		}
	}
	err = message.Wrap(
		err,
		message.EAuthOIDCTimeout,
		"Cannot authenticate at this time.",
		"Timeout during OIDC authentication flow.",
	)
	o.logger.Debug(err)
	return "", err
}
