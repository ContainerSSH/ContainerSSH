package auth

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"time"

    "go.containerssh.io/libcontainerssh/config"
    "go.containerssh.io/libcontainerssh/http"
    "go.containerssh.io/libcontainerssh/internal/structutils"
    "go.containerssh.io/libcontainerssh/log"
    "go.containerssh.io/libcontainerssh/message"
    "go.containerssh.io/libcontainerssh/metadata"
)

//region AuthGitHubConfig

//endregion

//region gitHubProvider

// newGitHubProvider creates a new, GitHub-specific OAuth2 provider.
func newGitHubProvider(cfg config.AuthOAuth2ClientConfig, logger log.Logger) (OAuth2Provider, error) {
	if cfg.Provider != config.AuthOAuth2GitHubProvider {
		return nil, fmt.Errorf("GitHub is not configured as the oAuth2 provider")
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid GitHub configuration (%w)", err)
	}

	parsedURL, err := url.Parse(cfg.GitHub.URL)
	if err != nil {
		return nil, message.Wrap(
			err,
			message.EAuthConfigError,
			"Failed to parse GitHub URL (%s)",
			cfg.GitHub.URL,
		)
	}

	parsedAPIURL, err := url.Parse(cfg.GitHub.APIURL)
	if err != nil {
		return nil, message.Wrap(
			err,
			message.EAuthConfigError,
			"Failed to parse GitHub API URL (%s)",
			cfg.GitHub.APIURL,
		)
	}

	wwwClientConfig := config.HTTPClientConfiguration{}
	structutils.Defaults(&wwwClientConfig)
	wwwClientConfig.URL = cfg.GitHub.URL
	wwwClientConfig.CACert = cfg.GitHub.CACert
	wwwClientConfig.Timeout = cfg.GitHub.RequestTimeout
	wwwClientConfig.RequestEncoding = http.RequestEncodingWWWURLEncoded
	if err := wwwClientConfig.Validate(); err != nil {
		return nil, err
	}

	jsonWWWClientConfig := config.HTTPClientConfiguration{}
	structutils.Defaults(&jsonWWWClientConfig)
	jsonWWWClientConfig.URL = cfg.GitHub.URL
	jsonWWWClientConfig.CACert = cfg.GitHub.CACert
	jsonWWWClientConfig.Timeout = cfg.GitHub.RequestTimeout
	jsonWWWClientConfig.RequestEncoding = http.RequestEncodingWWWURLEncoded
	if err := jsonWWWClientConfig.Validate(); err != nil {
		return nil, err
	}

	apiClientConfig := config.HTTPClientConfiguration{}
	structutils.Defaults(&apiClientConfig)
	apiClientConfig.URL = cfg.GitHub.APIURL
	apiClientConfig.CACert = cfg.GitHub.CACert
	apiClientConfig.Timeout = cfg.GitHub.RequestTimeout
	apiClientConfig.RequestEncoding = http.RequestEncodingWWWURLEncoded
	if err := apiClientConfig.Validate(); err != nil {
		return nil, err
	}

	return &gitHubProvider{
		logger:                logger,
		url:                   parsedURL,
		apiURL:                parsedAPIURL,
		clientID:              cfg.ClientID,
		clientSecret:          cfg.ClientSecret,
		requiredOrgMembership: cfg.GitHub.RequireOrgMembership,
		scopes:                cfg.GitHub.ExtraScopes,
		enforceUsername:       cfg.GitHub.EnforceUsername,
		enforceScopes:         cfg.GitHub.EnforceScopes,
		require2FA:            cfg.GitHub.Require2FA,
		wwwClientConfig:       wwwClientConfig,
		jsonWWWClientConfig:   jsonWWWClientConfig,
		apiClientConfig:       apiClientConfig,
	}, nil
}

type gitHubProvider struct {
	logger                log.Logger
	url                   *url.URL
	apiURL                *url.URL
	clientID              string
	clientSecret          string
	requiredOrgMembership string
	scopes                []string
	enforceScopes         bool
	require2FA            bool
	enforceUsername       bool
	wwwClientConfig       config.HTTPClientConfiguration
	jsonWWWClientConfig   config.HTTPClientConfiguration
	apiClientConfig       config.HTTPClientConfiguration
}

func (p *gitHubProvider) SupportsDeviceFlow() bool {
	return true
}

func (p *gitHubProvider) GetDeviceFlow(meta metadata.ConnectionAuthPendingMetadata) (OAuth2DeviceFlow, error) {
	flow, err := p.createFlow(meta)
	if err != nil {
		return nil, err
	}

	return &gitHubDeviceFlow{
		meta:       meta,
		gitHubFlow: flow,
		interval:   10 * time.Second,
	}, nil
}

func (p *gitHubProvider) SupportsAuthorizationCodeFlow() bool {
	return true
}

func (p *gitHubProvider) GetAuthorizationCodeFlow(meta metadata.ConnectionAuthPendingMetadata) (
	OAuth2AuthorizationCodeFlow,
	error,
) {
	flow, err := p.createFlow(meta)
	if err != nil {
		return nil, err
	}

	return &gitHubAuthorizationCodeFlow{
		meta:       meta,
		gitHubFlow: flow,
	}, nil
}

func (p *gitHubProvider) createFlow(meta metadata.ConnectionAuthPendingMetadata) (
	gitHubFlow,
	error,
) {
	logger := p.logger.WithLabel("connectionID", meta.ConnectionID).WithLabel("username", meta.Username)

	client, err := http.NewClient(p.wwwClientConfig, logger)
	if err != nil {
		return gitHubFlow{}, message.WrapUser(
			err,
			message.EAuthGitHubHTTPClientCreateFailed,
			"Authentication currently unavailable.",
			"Cannot create GitHub device flow authenticator because the HTTP client configuration failed.",
		)
	}

	jsonClient, err := http.NewClient(p.jsonWWWClientConfig, logger)
	if err != nil {
		return gitHubFlow{}, message.WrapUser(
			err,
			message.EAuthGitHubHTTPClientCreateFailed,
			"Authentication currently unavailable.",
			"Cannot create GitHub device flow authenticator because the HTTP client configuration failed.",
		)
	}

	flow := gitHubFlow{
		provider:        p,
		clientID:        p.clientID,
		clientSecret:    p.clientSecret,
		logger:          logger,
		client:          client,
		jsonClient:      jsonClient,
		apiClientConfig: p.apiClientConfig,
	}
	return flow, nil
}

func (p *gitHubProvider) getScope() string {
	scopes := p.scopes
	if p.requiredOrgMembership != "" {
		foundOrgRead := false
		for _, scope := range scopes {
			if scope == "org" || scope == "read:org" {
				foundOrgRead = true
				break
			}
		}
		if !foundOrgRead {
			scopes = append(scopes, "read:org")
		}
	}
	if p.require2FA {
		foundUserRead := false
		for _, scope := range scopes {
			if scope == "user" || scope == "read:user" {
				foundUserRead = true
				break
			}
		}
		if !foundUserRead {
			scopes = append(scopes, "read:user")
		}
	}
	return strings.Join(scopes, ",")
}

type gitHubDeleteAccessTokenRequest struct {
	AccessToken string `json:"access_token"`
}

type gitHubDeleteAccessTokenResponse struct {
}

type gitHubAccessTokenRequest struct {
	ClientID     string `json:"client_id" schema:"client_id,required"`
	ClientSecret string `json:"client_secret,omitempty" schema:"client_secret"`
	Code         string `json:"code,omitempty" schema:"code"`
	DeviceCode   string `json:"device_code,omitempty" schema:"device_code"`
	GrantType    string `json:"grant_type,omitempty" schema:"grant_type"`
	State        string `json:"state,omitempty" schema:"state"`
}

type gitHubAccessTokenResponse struct {
	AccessToken      string `json:"access_token,omitempty"`
	Scope            string `json:"scope,omitempty"`
	TokenType        string `json:"token_type,omitempty"`
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
	ErrorURI         string `json:"error_uri,omitempty"`
	Interval         uint   `json:"interval,omitempty"`
}

type gitHubUserResponse struct {
	Login                   string `json:"login"`
	ID                      uint64 `json:"id"`
	NodeID                  string `json:"node_id"`
	AvatarURL               string `json:"avatar_url"`
	ProfileURL              string `json:"html_url"`
	Name                    string `json:"name"`
	Company                 string `json:"company"`
	BlogURL                 string `json:"blog"`
	Location                string `json:"location"`
	Email                   string `json:"email"`
	Bio                     string `json:"bio"`
	TwitterUsername         string `json:"twitter_username"`
	TwoFactorAuthentication *bool  `json:"two_factor_authentication"`
}

// endregion

//region gitHubFlow

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
						"The username entered in your SSH client does not match your GitHub login.",
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
			"Cannot create GitHub device flow authenticator because the HTTP client configuration failed.",
		)
	}
	return apiClient, nil
}

func (g *gitHubFlow) Deauthorize(ctx context.Context) {
	if g.accessToken == "" {
		return
	}
loop:
	for {
		req := &gitHubDeleteAccessTokenRequest{
			AccessToken: g.accessToken,
		}
		apiClient, err := g.getAPIClient(g.accessToken, true)
		if err != nil {
			g.logger.Warning(message.Wrap(err,
				message.EAuthGitHubDeleteAccessTokenFailed, "Failed to delete access token"))
			return
		}
		statusCode, err := apiClient.Delete(
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

//endregion

//region gitHubAuthorizationCodeFlow

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
		message.EAuthGitHubTimeout,
		"Timeout while trying to obtain GitHub authentication data.",
		"Timeout while trying to obtain GitHub authentication data.",
	)
	g.logger.Debug(err)
	return "", err
}

//endregion

//region gitHubDeviceFlow

type gitHubDeviceFlow struct {
	gitHubFlow

	interval   time.Duration
	deviceCode string
	meta       metadata.ConnectionAuthPendingMetadata
}

func (d *gitHubDeviceFlow) GetAuthorizationURL(ctx context.Context) (
	verificationLink string,
	userCode string,
	expiresIn time.Duration,
	err error,
) {
	req := &gitHubDeviceRequest{
		ClientID: d.provider.clientID,
		Scope:    d.provider.getScope(),
	}
	var lastError error
	var statusCode int
loop:
	for {
		resp := &gitHubDeviceResponse{}
		statusCode, lastError = d.client.Post("/login/device/code", req, resp)
		if lastError == nil {
			if statusCode == 200 {
				d.interval = time.Duration(resp.Interval) * time.Second
				d.deviceCode = resp.DeviceCode
				return resp.VerificationURI, resp.UserCode, time.Duration(resp.ExpiresIn) * time.Second, nil
			} else {
				switch resp.Error {
				case "slow_down":
					// Let's assume this means that we reached the 50/hr limit. This is currently undocumented.
					lastError = message.UserMessage(
						message.EAuthGitHubDeviceAuthorizationLimit,
						"Cannot authenticate at this time.",
						"GitHub device authorization limit reached (%s).",
						resp.ErrorDescription,
					)
					d.logger.Debug(lastError)
					return "", "", 0, lastError
				}
			}
			lastError = message.UserMessage(
				message.EAuthGitHubDeviceCodeRequestFailed,
				"Cannot authenticate at this time.",
				"Non-200 status code from GitHub device code API (%d; %s; %s).",
				statusCode,
				resp.Error,
				resp.ErrorDescription,
			)
			d.logger.Debug(lastError)
		}
		d.logger.Debug(lastError)
		select {
		case <-time.After(10 * time.Second):
			continue
		case <-ctx.Done():
			break loop
		}
	}
	err = message.WrapUser(
		lastError,
		message.EAuthGitHubTimeout,
		"Cannot authenticate at this time.",
		"Timeout while trying to obtain a GitHub device code.",
	)
	d.logger.Debug(err)
	return "", "", 0, err
}

func (d *gitHubDeviceFlow) Verify(ctx context.Context) (string, metadata.ConnectionAuthenticatedMetadata, error) {
	accessToken, err := d.getAccessToken(ctx)
	d.accessToken = accessToken
	if err != nil {
		if accessToken != "" {
			d.Deauthorize(ctx)
		}
		return "", d.meta.AuthFailed(), err
	}
	return d.getIdentity(ctx, d.meta, accessToken)
}

func (d *gitHubDeviceFlow) getAccessToken(ctx context.Context) (string, error) {
	var statusCode int
	var lastError error
loop:
	for {
		req := &gitHubAccessTokenRequest{
			ClientID:   d.provider.clientID,
			DeviceCode: d.deviceCode,
			GrantType:  "urn:ietf:params:oauth:grant-type:device_code",
		}
		resp := &gitHubAccessTokenResponse{}
		statusCode, lastError = d.client.Post("/login/oauth/access_token", req, resp)
		if statusCode != 200 {
			if resp.Error == "authorization_pending" {
				lastError = message.NewMessage(
					message.EAuthGitHubAuthorizationPending,
					"User authorization still pending, retrying in %d seconds.",
					d.interval,
				)
			} else {
				lastError = message.UserMessage(
					message.EAuthGitHubAccessTokenFetchFailed,
					"Cannot authenticate at this time.",
					"Non-200 status code from GitHub access token API (%d; %s; %s).",
					statusCode,
					resp.Error,
					resp.ErrorDescription,
				)
			}
		} else if lastError == nil {
			switch resp.Error {
			case "authorization_pending":
				lastError = message.UserMessage(message.EAuthGitHubAuthorizationPending, "Authentication is still pending.", "The user hasn't completed the authentication process.")
			case "slow_down":
				if resp.Interval > 15 {
					// Assume we have exceeded the hourly rate limit, let's fall back.
					return "", message.UserMessage(message.EAuthDeviceFlowRateLimitExceeded, "Cannot authenticate at this time. Please try again later.", "Rate limit for device flow exceeded, attempting authorization code flow.")
				}
			case "expired_token":
				return "", fmt.Errorf("BUG: expired token during device flow authentication")
			case "unsupported_grant_type":
				return "", fmt.Errorf("BUG: unsupported grant type error while trying device authorization")
			case "incorrect_client_credentials":
				// User entered the incorrect device code
				return "", message.UserMessage(message.EAuthIncorrectClientCredentials, "GitHub authentication failed", "User entered incorrect device code")
			case "incorrect_device_code":
				// User entered the incorrect device code
				return "", message.UserMessage(message.EAuthFailed, "GitHub authentication failed", "User entered incorrect device code")
			case "access_denied":
				// User hit don't authorize
				return "", message.UserMessage(message.EAuthFailed, "GitHub authentication failed", "User canceled GitHub authentication")
			case "":
				return resp.AccessToken, d.checkGrantedScopes(resp.Scope)
			}
		}
		d.logger.Debug(lastError)
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(d.interval):
		}
	}
	err := message.WrapUser(
		lastError,
		message.EAuthGitHubTimeout,
		"Timeout while trying to obtain GitHub authentication data.",
		"Timeout while trying to obtain GitHub authentication data.",
	)
	d.logger.Debug(err)
	return "", err
}

type gitHubDeviceRequest struct {
	ClientID string `schema:"client_id"`
	Scope    string `schema:"scope"`
}

type gitHubDeviceResponse struct {
	DeviceCode       string `json:"device_code"`
	UserCode         string `json:"user_code"`
	VerificationURI  string `json:"verification_uri"`
	ExpiresIn        uint   `json:"expires_in" yaml:"expires_in"`
	Interval         uint   `json:"interval" yaml:"interval"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	ErrorURI         string `json:"error_uri"`
}

//endregion
