package auth

import (
	"github.com/containerssh/libcontainerssh/metadata"
)

// PasswordAuthRequest is an authentication request for password authentication.
//
// swagger:model PasswordAuthRequest
type PasswordAuthRequest struct {
	metadata.ConnectionAuthPendingMetadata `json:",inline"`

	// Password the user provided for authentication.
	//
	// required: true
	// swagger:strfmt Base64
	Password string `json:"passwordBase64"`
}

// PublicKeyAuthRequest is an authentication request for public key authentication.
//
// swagger:model PublicKeyAuthRequest
type PublicKeyAuthRequest struct {
	metadata.ConnectionAuthPendingMetadata `json:",inline"`

	PublicKey `json:",inline"`
}

// AuthorizationRequest is the authorization request used after some
// authentication methods (e.g. kerberos) to determine whether users are
// allowed to access the service
//
// swagger:model AuthorizationRequest
type AuthorizationRequest struct {
	metadata.ConnectionAuthenticatedMetadata `json:",inline"`
}

// ResponseBody is a response to authentication requests.
//
// swagger:model AuthResponseBody
type ResponseBody struct {
	metadata.ConnectionAuthenticatedMetadata `json:",inline"`

	// Success indicates if the authentication was successful.
	//
	// required: true
	// in: body
	Success bool `json:"success"`
}

// Response is the full HTTP authentication response.
//
// swagger:response AuthResponse
type Response struct {
	// The response body
	//
	// in: body
	ResponseBody
}
