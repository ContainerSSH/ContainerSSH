package auth

import (
	"context"
	"fmt"
	"time"

	"go.containerssh.io/libcontainerssh/message"
	"go.containerssh.io/libcontainerssh/metadata"
)

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
				message.EAuthOAuth2DeviceCodeRequestFailed,
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
		message.EAuthOAuth2Timeout,
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
					message.EAuthOAuth2AuthorizationPending,
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
				lastError = message.UserMessage(message.EAuthOAuth2AuthorizationPending, "Authentication is still pending.", "The user hasn't completed the authentication process.")
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
		message.EAuthOAuth2Timeout,
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
