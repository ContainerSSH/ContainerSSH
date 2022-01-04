package auth

// PasswordAuthRequest is an authentication request for password authentication.
//
// swagger:model PasswordAuthRequest
type PasswordAuthRequest struct {
	// Username is the username provided for authentication.
	//
	// required: true
	Username string `json:"username"`

	// RemoteAddress is the IP address of the user trying to authenticate.
	//
	// required: true
	RemoteAddress string `json:"remoteAddress"`

	// ConnectionID is an opaque ID to identify the SSH connection in question.
	//
	// required: true
	ConnectionID string `json:"connectionId"`

	// SessionID is a deprecated alias for ConnectionID and will be removed in the future.
	//
	// required: true
	SessionID string `json:"sessionId"`

	// Password the user provided for authentication.
	//
	// required: true
	Password string `json:"passwordBase64"`
}

// PublicKeyAuthRequest is an authentication request for public key authentication.
//
// swagger:model PublicKeyAuthRequest
type PublicKeyAuthRequest struct {
	// Username is the username provided for authentication.
	//
	// required: true
	Username string `json:"username"`

	// RemoteAddress is the IP address of the user trying to authenticate.
	//
	// required: true
	RemoteAddress string `json:"remoteAddress"`

	// ConnectionID is an opaque ID to identify the SSH connection in question.
	//
	// required: true
	ConnectionID string `json:"connectionId"`

	// SessionID is a deprecated alias for ConnectionID and will be removed in the future.
	//
	// required: true
	SessionID string `json:"sessionId"`

	// PublicKey is the key in the authorized key format.
	//
	// required: true
	PublicKey string `json:"publicKey"`
}

// AuthorizationRequest is the authorization request used after some
// authentication methods (e.g. kerberos) to determine whether users are
// allowed to access the service
//
// swagger:model AuthorizationRequest
type AuthorizationRequest struct {
	// PrincipalUsername is the authenticated username of the user, this
	// username has been verified to be correct and to correspond to the
	// user that is connecting.
	//
	// required: true
	PrincipalUsername string `json:"principalUsername"`
	// LoginUsername is the username the user wishes to log in as. In
	// general, the authorization must check that this matches the
	// PrincipalUsername, however in some cases it may be beneficial to let
	// some users log in as others (e.g. administrators logging in as
	// normal users to debug)
	//
	// required: true
	LoginUsername string `json:"loginUsername"`
	// RemoteAddress is the address the user is connecting from
	//
	// required: true
	RemoteAddress string `json:"remoteAddress"`
	// ConnectionID is an opaque ID to identify the SSH connection
	//
	// required: true
	ConnectionID string `json:"connectionId"`
}

// ResponseBody is a response to authentication requests.
//
// swagger:model AuthResponseBody
type ResponseBody struct {
	// Success indicates if the authentication was successful.
	//
	// required: true
	Success bool `json:"success"`

	// Metadata is a set of key-value pairs that can be returned and either
	// consumed by the configuration server or
	// exposed in the backend as environment variables.
	// They can also be used to deploy files in the container
	//
	// required: false
	Metadata *ConnectionMetadata `json:"metadata,omitempty"`
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
