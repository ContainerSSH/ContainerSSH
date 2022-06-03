package webhook

import (
    auth2 "go.containerssh.io/libcontainerssh/auth"
    "go.containerssh.io/libcontainerssh/config"
    "go.containerssh.io/libcontainerssh/internal/auth"
    "go.containerssh.io/libcontainerssh/internal/geoip/dummy"
    "go.containerssh.io/libcontainerssh/internal/metrics"
    "go.containerssh.io/libcontainerssh/log"
    "go.containerssh.io/libcontainerssh/metadata"
)

type Client interface {
	// Password authenticates with a password from the client. It returns a bool if the authentication as successful
	// or not. If an error happened while contacting the authentication server it will return an error.
	Password(
		metadata metadata.ConnectionAuthPendingMetadata,
		password []byte,
	) AuthenticationResponse

	// PubKey authenticates with a public key from the client. It returns a bool if the authentication as successful
	// or not. If an error happened while contacting the authentication server it will return an error.
	PubKey(
		metadata metadata.ConnectionAuthPendingMetadata,
		pubKey auth2.PublicKey,
	) AuthenticationResponse
}

// AuthenticationResponse holds the results of an authentication.
type AuthenticationResponse interface {
	// Success must return true or false of the authentication was successful / unsuccessful.
	Success() bool
	// Error returns the error that happened during the authentication. This is useful for returning detailed error
	// message.
	Error() error
	// Metadata returns a set of metadata entries that have been obtained during the authentication.
	Metadata() metadata.ConnectionAuthenticatedMetadata
}

// NewTestClient creates a new copy of a client usable for testing purposes.
func NewTestClient(cfg config.AuthWebhookClientConfig, logger log.Logger) (Client, error) {
	metricsCollector := metrics.New(dummy.New())

	authClient, err := auth.NewWebhookClient(
		auth.AuthenticationTypeAll,
		cfg,
		logger,
		metricsCollector,
	)
	if err != nil {
		return nil, err
	}
	return &authClientWrapper{
		authClient,
	}, nil
}

type authClientWrapper struct {
	c auth.WebhookClient
}

func (a authClientWrapper) Password(
	metadata metadata.ConnectionAuthPendingMetadata,
	password []byte,
) AuthenticationResponse {
	return a.c.Password(metadata, password)
}

func (a authClientWrapper) PubKey(
	metadata metadata.ConnectionAuthPendingMetadata,
	pubKey auth2.PublicKey,
) AuthenticationResponse {
	return a.c.PubKey(metadata, pubKey)
}
