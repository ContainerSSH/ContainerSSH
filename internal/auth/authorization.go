// This file contains the interface definitions for the authorization call.

package auth

import (
	"go.containerssh.io/libcontainerssh/metadata"
)

// AuthzProvider provides a method to verify the authenticated username against the username provided by the user.
// These two can differ when the authentication server verified the credentials based on other factors, for example
// using oAuth2, OIDC, or Kerberos.
type AuthzProvider interface {
	// Authorize checks the username from the authentication against the username provided by the user. These two can
	// differ in case of oAuth2, OIDC, or Kerberos authentication. This can be used for an administrator to log in as
	// a different user. The function SHOULD reject the user if the combination of provided and authenticated username
	// is invalid, or the user is not eligible to access this system.
	Authorize(
		metadata metadata.ConnectionAuthenticatedMetadata,
	) AuthorizationResponse
}

// AuthorizationResponse is the result of the authorization process in AuthzProvider.
type AuthorizationResponse interface {
	// Success must return true or false of the authorization was successful / unsuccessful.
	Success() bool
	// Error returns the error that happened during the authorization.
	Error() error
	// Metadata returns a set of metadata entries that have been returned from the authorization process. The values
	// provided will be merged with the values obtained from the authentication process, where the authorization
	// values have precedence. If a value is not returned in this field the value from the authentication process is
	// taken.
	Metadata() metadata.ConnectionAuthenticatedMetadata
	// OnDisconnect is called when the urlEncodedClient disconnects, or if the authorization fails due to a different reason.
	// This hook can be used to clean up issued temporary credentials.
	OnDisconnect()
}
