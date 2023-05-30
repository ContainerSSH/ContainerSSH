package auth

import (
	"context"
	"strings"
	"time"

	"go.containerssh.io/libcontainerssh/config"
	http2 "go.containerssh.io/libcontainerssh/http"
	"go.containerssh.io/libcontainerssh/log"
	"go.containerssh.io/libcontainerssh/message"
	"go.containerssh.io/libcontainerssh/metadata"
)

type oidcFlow struct {
	provider          *oidcProvider
	connectionID      string
	username          string
	logger            log.Logger
	urlEncodedClient  http2.Client
	meta              metadata.ConnectionAuthPendingMetadata
	discoveryResponse oidcDiscoveryResponse
	accessToken       string
}

type revocationRequest struct {
	Token         string `json:"token" schema:"token"`
	TokenTypeHint string `json:"token_type_hint" schema:"token_type_hint"`
}

func (o *oidcFlow) getAuthenticatedHTTPClient(accessToken string) (http2.Client, error) {
	cfg := o.provider.config.HTTPClientConfiguration
	cfg.RequestEncoding = config.RequestEncodingWWWURLEncoded
	return http2.NewClientWithHeaders(
		cfg,
		o.logger,
		map[string][]string{
			"authorization": {
				"Bearer " + accessToken,
			},
		},
		true,
	)
}

type deauthorizeResponse struct {
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
}

func (o *oidcFlow) Deauthorize(ctx context.Context) {
	defer func() {
		o.accessToken = ""
	}()
	if o.accessToken == "" {
		return
	}
	requestBody := &revocationRequest{
		Token:         o.accessToken,
		TokenTypeHint: "access_token",
	}
	var response deauthorizeResponse
	for {
		statusCode, err := o.urlEncodedClient.RequestURL(
			"POST",
			o.discoveryResponse.RevocationEndpoint,
			requestBody,
			&response,
		)
		if err == nil && response.Error != "" {
			if response.Error == "invalid_token" {
				return
			}
			err = message.NewMessage(
				message.EAuthOIDCDeauthorizeFailed,
				"Failed to revoke access token (%s: %s)",
				response.Error,
				response.ErrorDescription,
			)
		}
		// If the revocation is successful the return data is not json, so we expect the decoding to fail. Therefore ignore errors if statusCode==200
		if err == nil || statusCode == 200 {
			return
		}
		err = message.Wrap(
			err,
			message.EAuthOIDCDeauthorizeFailed,
			"Failed to revoke access token.",
		)
		o.logger.Debug(err)
		if statusCode > 399 && statusCode < 500 {
			return
		}

		select {
		case <-time.After(5 * time.Second):
		case <-ctx.Done():
			err = message.Wrap(
				err,
				message.EAuthOIDCDeauthorizeTimeout,
				"Timeout while trying to revoke authorization token.",
			)
			o.logger.Warning(err)
			return
		}
	}
}

func (o *oidcFlow) getIdentity(
	ctx context.Context,
	meta metadata.ConnectionAuthPendingMetadata,
	token string,
) (string, metadata.ConnectionAuthenticatedMetadata, error) {
	client, err := o.getAuthenticatedHTTPClient(token)
	if err != nil {
		return "", meta.AuthFailed(), err
	}

	for {
		resp := map[string]interface{}{}
		statusCode, err := client.RequestURL("GET", o.discoveryResponse.UserInfoEndpoint, nil, &resp)
		if err == nil {
			if statusCode > 100 && statusCode < 300 {
				username, ok := resp[o.provider.config.UsernameField]
				if !ok {
					err = message.UserMessage(
						message.EAuthOIDCNoUsername,
						"Authentication currently unavailable",
						"The OIDC server did not return the %s field in the response, which was configured to "+
							"be used as the username. Please check your ContainerSSH configuration against your OIDC "+
							"server configuration.",
						o.provider.config.UsernameField,
					)
					o.logger.Error(err)
					return token, meta.AuthFailed(), err
				}
				usernameString, ok := username.(string)
				if !ok {
					err = message.UserMessage(
						message.EAuthOIDCNoUsername,
						"Authentication currently unavailable",
						"The OIDC server returned the %s field in the response, which was configured to "+
							"be used as the username, but it was not a string.",
						o.provider.config.UsernameField,
					)
					o.logger.Error(err)
					return token, meta.AuthFailed(), err
				}
				m := meta.GetMetadata()
				m["OIDC_TOKEN"] = metadata.Value{Value: token, Sensitive: true}
				for field, value := range resp {
					valueString, ok := value.(string)
					if ok {
						m["OIDC_USERINFO_"+strings.ToUpper(field)] = metadata.Value{
							Value: valueString,
						}
					}
				}
				return token, meta.Authenticated(usernameString), nil
			}
			err = message.UserMessage(
				message.EAuthOIDCUserInfoFetchFailed,
				"Authentication currently unavailable",
				"Non-200 status code from OIDC userinfo API (%d).",
				statusCode,
			)
		}
		o.logger.Debug(err)

		select {
		case <-time.After(5 * time.Second):
		case <-ctx.Done():
			err = message.Wrap(
				err,
				message.EAuthOIDCTimeout,
				"Timeout while trying to get user info.",
			)
			o.logger.Warning(err)
			return "", meta.AuthFailed(), err
		}
	}
}
