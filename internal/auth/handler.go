package auth

import (
	auth2 "go.containerssh.io/libcontainerssh/auth"
	"go.containerssh.io/libcontainerssh/metadata"
)

type Handler interface {
	// OnPassword is called if the urlEncodedClient requests a password authentication.
	//
	// - meta is the metadata of the connection, including the username provided by the user.
	// - password is the password the user entered.
	//
	// The method must return a boolean if the authentication was successful, and an error if the authentication failed
	// for other reasons (e.g. backend database was not available). If an error is returned the server responds with
	// an HTTP 500 response.
	OnPassword(
		metadata metadata.ConnectionAuthPendingMetadata,
		password []byte,
	) (bool, metadata.ConnectionAuthenticatedMetadata, error)

	// OnPubKey is called when the urlEncodedClient requests a public key authentication.
	//
	// - meta is the metadata of the connection, including the username provided by the user.
	// - publicKey contains the SSH public key and other accompanying information about the key.
	//
	// The method must return a boolean if the authentication was successful, and an error if the authentication failed
	// for other reasons (e.g. backend database was not available). If an error is returned the server responds with
	// an HTTP 500 response.
	OnPubKey(
		meta metadata.ConnectionAuthPendingMetadata,
		publicKey auth2.PublicKey,
	) (bool, metadata.ConnectionAuthenticatedMetadata, error)

	// OnAuthorization is called when the urlEncodedClient requests user authorization.
	//
	// - meta contains the username the user provided, the authenticated username, and other information about
	//   the connecting user.
	OnAuthorization(
		meta metadata.ConnectionAuthenticatedMetadata,
	) (bool, metadata.ConnectionAuthenticatedMetadata, error)
}
