package auth

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"go.containerssh.io/libcontainerssh/config"
	"go.containerssh.io/libcontainerssh/http"
	"go.containerssh.io/libcontainerssh/log"
	"go.containerssh.io/libcontainerssh/message"
	"go.containerssh.io/libcontainerssh/metadata"
)

type gitHubFlow struct {
	provider        *gitHubProvider
	accessToken     string
	clientID        string
	clientSecret    string
	logger          log.Logger
	client          http.Client
	jsonClient      http.Client
	apiClientConfig config.HTTPClientConfiguration
}

func (g *gitHubFlow) checkGrantedScopes(scope string) error {
	grantedScopes := strings.Split(scope, ",")
	if g.provider.enforceScopes {
		for _, requiredScope := range g.provider.scopes {
			scopeGranted := false
			requiredScopeParts := strings.Split(requiredScope, ":")
			for _, grantedScope := range grantedScopes {
				if grantedScope == requiredScope || (len(requiredScopeParts) > 1 && requiredScopeParts[0] == grantedScope) {
					scopeGranted = true
					break
				}
			}
			if !scopeGranted {
				err := message.UserMessage(
					message.EAuthGitHubRequiredScopeNotGranted,
					fmt.Sprintf("You have not granted us the required %s permission.", requiredScope),
					"The user has not granted the %s permission.",
					requiredScope,
				)
				g.logger.Debug(err)
				return err
			}
		}
	}
	if g.provider.requiredOrgMembership != "" {
		for _, grantedScope := range grantedScopes {
			if grantedScope == "org" || grantedScope == "read:org" {
				return nil
			}
		}
		err := message.UserMessage(
			message.EAuthGitHubRequiredScopeNotGranted,
			"You have not granted us permissions to read your organization memberships required for login.",
			"The user has not granted the org or read:org memberships required to validate the organization member ship.",
		)
		g.logger.Debug(err)
		return err
	}
	return nil
}

func (g *gitHubFlow) getIdentity(
	ctx context.Context,
	meta metadata.ConnectionAuthPendingMetadata,
	accessToken string,
) (string, metadata.ConnectionAuthenticatedMetadata, error) {
	var statusCode int
	var lastError error
	apiClient, err := g.getAPIClient(accessToken, false)
	if err != nil {
		return g.accessToken, meta.AuthFailed(), err
	}
loop:
	for {
		response := &gitHubUserResponse{}
		statusCode, lastError = apiClient.Get("/user", response)
		if lastError == nil {
			if statusCode == 200 {
				if g.provider.enforceUsername && response.Login != meta.Username {
					err := message.UserMessage(
						message.EAuthUsernameDoesNotMatch,
						"The username entered in your SSH urlEncodedClient does not match your GitHub login.",
						"The user's username entered in the SSH username and on GitHub login do not match, but enforceUsername is enabled.",
					)
					g.logger.Debug(err)
					return g.accessToken, meta.AuthFailed(), err
				}

				result := map[string]string{}
				if response.TwoFactorAuthentication != nil {
					if *response.TwoFactorAuthentication {
						result["GITHUB_2FA"] = "true"
					} else {
						if g.provider.require2FA {
							err := message.UserMessage(
								message.EAuthGitHubNo2FA,
								"Please enable two-factor authentication on GitHub to access this server.",
								"The user does not have two-factor authentication enabled on their GitHub account.",
							)
							g.logger.Debug(err)
							return g.accessToken, meta.AuthFailed(), err
						}
						result["GITHUB_2FA"] = "false"
					}
				} else if g.provider.require2FA {
					err := message.UserMessage(
						message.EAuthGitHubNo2FA,
						"Please grant the read:user permission so we can check your 2FA status.",
						"The user did not provide the read:user permission to read the 2FA status.",
					)
					g.logger.Debug(err)
					return g.accessToken, meta.AuthFailed(), err
				}
				m := meta.GetMetadata()
				m["GITHUB_METHOD"] = metadata.Value{Value: "device"}
				// Note: we are adding all entries as sensitive since they are personally identifiable data and should
				// not be passed around or logged needlessly. Doing so would possibly incur a problem under GDPR.
				m["GITHUB_TOKEN"] = metadata.Value{Value: accessToken, Sensitive: true}
				m["GITHUB_LOGIN"] = metadata.Value{Value: response.Login, Sensitive: true}
				m["GITHUB_ID"] = metadata.Value{Value: fmt.Sprintf("%d", response.ID), Sensitive: true}
				m["GITHUB_NODE_ID"] = metadata.Value{Value: response.NodeID, Sensitive: true}
				m["GITHUB_NAME"] = metadata.Value{Value: response.Name, Sensitive: true}
				m["GITHUB_AVATAR_URL"] = metadata.Value{Value: response.AvatarURL, Sensitive: true}
				m["GITHUB_BIO"] = metadata.Value{Value: response.Bio, Sensitive: true}
				m["GITHUB_COMPANY"] = metadata.Value{Value: response.Company, Sensitive: true}
				m["GITHUB_EMAIL"] = metadata.Value{Value: response.Email, Sensitive: true}
				m["GITHUB_BLOG_URL"] = metadata.Value{Value: response.BlogURL, Sensitive: true}
				m["GITHUB_LOCATION"] = metadata.Value{Value: response.Location, Sensitive: true}
				m["GITHUB_TWITTER_USERNAME"] = metadata.Value{Value: response.TwitterUsername, Sensitive: true}
				m["GITHUB_PROFILE_URL"] = metadata.Value{Value: response.ProfileURL, Sensitive: true}
				m["GITHUB_AVATAR_URL"] = metadata.Value{Value: response.AvatarURL, Sensitive: true}
				if g.provider.enforceUsername && response.Login != meta.Username {
					return g.accessToken, meta.AuthFailed(), message.UserMessage(
						message.EAuthGitHubUsernameDoesNotMatch,
						"Your GitHub username does not match your SSH login. Please try again and specify your GitHub username when connecting.",
						"User did not use their GitHub username in the SSH login.",
					)
				}
				return g.accessToken, meta.Authenticated(response.Login), nil
			} else {
				g.logger.Debug(
					message.NewMessage(
						message.EAuthGitHubUserRequestFailed,
						"Request to GitHub user endpoint failed, non-200 response code (%d), retrying in 10 seconds...",
						statusCode,
					),
				)
			}
		} else {
			g.logger.Debug(
				message.Wrap(
					lastError,
					message.EAuthGitHubUserRequestFailed,
					"Request to GitHub user endpoint failed, retrying in 10 seconds...",
				),
			)
		}
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(10 * time.Second):
		}
	}
	err = message.WrapUser(
		lastError,
		message.EAuthGitHubUserRequestFailed,
		"Timeout while trying to fetch your identity from GitHub.",
		"Timeout while trying fetch user identity from GitHub.",
	)
	g.logger.Debug(err)
	return g.accessToken, meta.AuthFailed(), err
}

func (g *gitHubFlow) getAPIClient(token string, basicAuth bool) (http.Client, error) {
	headers := map[string][]string{}
	if basicAuth {
		headers["authorization"] = []string{
			fmt.Sprintf("basic %s", base64.StdEncoding.EncodeToString(
				[]byte(fmt.Sprintf(
					"%s:%s",
					g.clientID,
					g.clientSecret,
				)),
			)),
		}
	} else if token != "" {
		headers["authorization"] = []string{
			fmt.Sprintf("bearer %s", token),
		}
	}
	apiClient, err := http.NewClientWithHeaders(g.apiClientConfig, g.logger, headers, true)
	if err != nil {
		return nil, message.WrapUser(
			err,
			message.EAuthGitHubHTTPClientCreateFailed,
			"Authentication currently unavailable.",
			"Cannot create GitHub authenticator because the HTTP urlEncodedClient configuration failed.",
		)
	}
	return apiClient, nil
}

func (g *gitHubFlow) Deauthorize(ctx context.Context) {
	if g.accessToken == "" {
		return
	}
	var apiClient http.Client
	var err error
loop:
	for {
		req := &gitHubDeleteAccessTokenRequest{
			AccessToken: g.accessToken,
		}
		apiClient, err = g.getAPIClient(g.accessToken, true)
		if err != nil {
			g.logger.Warning(message.Wrap(err,
				message.EAuthGitHubDeleteAccessTokenFailed, "Failed to delete access token"))
			return
		}
		var statusCode int
		statusCode, err = apiClient.Delete(
			fmt.Sprintf("/applications/%s/token", g.clientID),
			req,
			nil,
		)
		if err == nil && statusCode == 204 {
			g.accessToken = ""
			return
		}
		if err != nil {
			g.logger.Debug(
				message.Wrap(
					err,
					message.EAuthGitHubDeleteAccessTokenFailed,
					"Failed to delete access token.",
				),
			)
		} else {
			g.logger.Debug(
				message.NewMessage(
					message.EAuthGitHubDeleteAccessTokenFailed,
					"Failed to delete access token, invalid status code: %d",
					statusCode,
				),
			)
		}
		select {
		case <-time.After(10 * time.Second):
		case <-ctx.Done():
			break loop
		}
	}
}
