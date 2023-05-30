package auth

import (
	"context"
	"net/url"
	"time"

	"go.containerssh.io/libcontainerssh/message"
	"go.containerssh.io/libcontainerssh/metadata"
)

type gitHubAuthorizationCodeFlow struct {
	gitHubFlow
	meta metadata.ConnectionAuthPendingMetadata
}

func (g *gitHubAuthorizationCodeFlow) GetAuthorizationURL(_ context.Context) (string, error) {
	var link = &url.URL{}
	*link = *g.provider.url
	link.Path = "/login/oauth/authorize"
	query := link.Query()
	query.Set("client_id", g.provider.clientID)
	query.Set("login", g.meta.Username)
	query.Set("scope", g.provider.getScope())
	query.Set("state", g.meta.ConnectionID)
	link.RawQuery = query.Encode()
	return link.String(), nil
}

func (g *gitHubAuthorizationCodeFlow) Verify(ctx context.Context, state string, authorizationCode string) (
	string,
	metadata.ConnectionAuthenticatedMetadata,
	error,
) {
	if state != g.meta.ConnectionID {
		return g.accessToken, g.meta.AuthFailed(), message.UserMessage(
			message.EAuthGitHubStateDoesNotMatch,
			"The returned code is invalid.",
			"The user provided a code that contained an invalid state component.",
		)
	}
	accessToken, err := g.getAccessToken(ctx, authorizationCode)
	g.accessToken = accessToken
	if err != nil {
		if accessToken != "" {
			g.Deauthorize(ctx)
		}
		return g.accessToken, g.meta.AuthFailed(), err
	}
	return g.getIdentity(ctx, g.meta, accessToken)
}

func (g *gitHubAuthorizationCodeFlow) getAccessToken(ctx context.Context, code string) (string, error) {
	var statusCode int
	var lastError error
loop:
	for {
		req := &gitHubAccessTokenRequest{
			ClientID:     g.provider.clientID,
			ClientSecret: g.provider.clientSecret,
			Code:         code,
			State:        g.meta.ConnectionID,
		}
		resp := &gitHubAccessTokenResponse{}
		statusCode, lastError = g.client.Post("/login/oauth/access_token", req, resp)
		if statusCode != 200 {
			lastError = message.UserMessage(
				message.EAuthGitHubAccessTokenFetchFailed,
				"Cannot authenticate at this time.",
				"Non-200 status code from GitHub access token API (%d; %s; %s).",
				statusCode,
				resp.Error,
				resp.ErrorDescription,
			)
		} else if lastError == nil {
			return resp.AccessToken, g.checkGrantedScopes(resp.Scope)
		}
		g.logger.Debug(lastError)
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(10 * time.Second):
		}
	}
	err := message.WrapUser(
		lastError,
		message.EAuthOAuth2Timeout,
		"Timeout while trying to obtain GitHub authentication data.",
		"Timeout while trying to obtain GitHub authentication data.",
	)
	g.logger.Debug(err)
	return "", err
}
