package auth

import (
	"context"
	"time"
)

// OAuth2Provider is an OAuth2 backing provider with a specific API implementation.
type OAuth2Provider interface {
	// SupportsDeviceFlow returns true if the provider supports authenticating via the device flow.
	SupportsDeviceFlow() bool
	// GetDeviceFlow returns the OAuth2DeviceFlow for a single client used for performing a device flow
	// authorization with the OAuth2 server. The method must panic if the device flow is not supported.
	GetDeviceFlow(connectionID string, username string) (OAuth2DeviceFlow, error)

	// SupportsAuthorizationCodeFlow returns true if the provider supports the authorization code flow.
	SupportsAuthorizationCodeFlow() bool
	// GetAuthorizationCodeFlow returns the OAuth2AuthorizationCodeFlow for a single client used for performing
	// authorization code flow authorization with the OAuth2 server. The method must panic if the device flow is not
	// supported.
	GetAuthorizationCodeFlow(connectionID string, username string) (OAuth2AuthorizationCodeFlow, error)
}

type OAuth2Flow interface {
	// Deauthorize contacts the OAuth2 server to deauthorize an access token.
	Deauthorize(ctx context.Context)
}

type OAuth2AuthorizationCodeFlow interface {
	OAuth2Flow

	// GetAuthorizationURL returns the authorization URL a user should be redirected to to begin the login process.
	GetAuthorizationURL(ctx context.Context) (string, error)

	// Verify verifies the authorizationCode with the OAuth2 server and obtains an access token. It can optionally also
	// a key-value map of parameters that are passed to the configuration server.
	//
	// The implementation should retry obtaining the access token until ctx is canceled if the server
	// responds in an unexpected fashion.
	Verify(ctx context.Context, state string, authorizationCode string) (map[string]string, error)
}

type OAuth2DeviceFlow interface {
	OAuth2Flow

	// GetAuthorizationURL returns the authorization URL a user should be redirected to to begin the login process.
	GetAuthorizationURL(ctx context.Context) (
		verificationLink string,
		userCode string,
		expiration time.Duration,
		err error,
	)

	// Verify starts polling the OAuth2 server if the authorization has been completed. The ctx parameter contains a
	// context that will be canceled when the client disconnects.
	//
	// This method should return the access token, and additionally may also return a key-value map of parameters
	// to be passed to the configuration server.
	//
	// The method should return when the OAuth2 server either returns with a positive or negative
	// result. It should not return if the final result has not been obtained yet.
	//
	// The implementation should return an error only if the
	Verify(ctx context.Context) (map[string]string, error)
}