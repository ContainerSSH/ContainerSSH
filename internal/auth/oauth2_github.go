package auth

import (
	"context"
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

func (p *gitHubProvider) GetDeviceFlow(ctx context.Context, meta metadata.ConnectionAuthPendingMetadata) (OAuth2DeviceFlow, error) {
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

func (p *gitHubProvider) GetAuthorizationCodeFlow(ctx context.Context, meta metadata.ConnectionAuthPendingMetadata) (OAuth2AuthorizationCodeFlow, error) {
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
			"Cannot create GitHub device flow authenticator because the HTTP urlEncodedClient configuration failed.",
		)
	}

	jsonClient, err := http.NewClient(p.jsonWWWClientConfig, logger)
	if err != nil {
		return gitHubFlow{}, message.WrapUser(
			err,
			message.EAuthGitHubHTTPClientCreateFailed,
			"Authentication currently unavailable.",
			"Cannot create GitHub device flow authenticator because the HTTP urlEncodedClient configuration failed.",
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
