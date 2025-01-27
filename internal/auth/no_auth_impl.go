package auth

import (
	"go.containerssh.io/containerssh/metadata"
)

type noneAuthenticator struct {
	enabled bool
}

func (a *noneAuthenticator) Context(
	meta metadata.ConnectionAuthPendingMetadata,
) AuthenticationContext {
	return &noneAuthenticationContext{
		enabled: a.enabled,
		metadata: meta,
	}
}

type noneAuthenticationContext struct {
	enabled  bool
	metadata metadata.ConnectionAuthPendingMetadata
}

// simply returns whether the no auth mechanism is enabled
func (a *noneAuthenticationContext) Success() bool {
	return a.enabled
}

// can never error
func (a *noneAuthenticationContext) Error() error {
	return nil
}
func (a *noneAuthenticationContext) Metadata() metadata.ConnectionAuthenticatedMetadata {
	return metadata.ConnectionAuthenticatedMetadata{
		ConnectionAuthPendingMetadata: a.metadata,
	}
}
func (a *noneAuthenticationContext) OnDisconnect() {}
