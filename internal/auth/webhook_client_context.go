package auth

import (
	"go.containerssh.io/libcontainerssh/metadata"
)

type webhookClientContext struct {
	meta    metadata.ConnectionAuthenticatedMetadata
	success bool
	err     error
}

func (h webhookClientContext) AuthenticatedUsername() string {
	return h.meta.AuthenticatedUsername
}

func (h webhookClientContext) Success() bool {
	return h.success
}

func (h webhookClientContext) Error() error {
	return h.err
}

func (h webhookClientContext) Metadata() metadata.ConnectionAuthenticatedMetadata {
	return h.meta
}

func (h webhookClientContext) OnDisconnect() {
}
