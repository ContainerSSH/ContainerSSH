package protocol

// Password authentication request.
//
// swagger:model PasswordAuthRequest
type PasswordAuthRequest struct {
	// The username provided for authentication (deprecated, use 'username' instead)
	//
	// required: true
	User string `json:"user"`
	// The username provided for authentication
	//
	// required: true
	Username string `json:"username"`
	// IP address of the user trying to authenticate
	//
	// required: true
	RemoteAddress string `json:"remoteAddress"`
	// Opaque ID to identify the SSH session in question
	//
	// required: true
	SessionId string `json:"sessionId"`
	// Password the user provided for authentication
	//
	// required: true
	Password string `json:"passwordBase64"`
}

// Public key authentication request.
//
// swagger:model PublicKeyAuthRequest
type PublicKeyAuthRequest struct {
	// The username provided for authentication (deprecated, use 'username' instead)
	//
	// required: true
	User string `json:"user"`
	// The username provided for authentication
	//
	// required: true
	Username string `json:"username"`
	// IP address of the user trying to authenticate
	//
	// required: true
	RemoteAddress string `json:"remoteAddress"`
	// Opaque ID to identify the SSH session in question
	//
	// required: true
	SessionId string `json:"sessionId"`
	// Serialized key data in SSH wire format
	//
	// required: true
	PublicKey string `json:"publicKeyBase64"`
}

// Response to authentication requests.
//
// swagger:model AuthResponseBody
type AuthResponse struct {
	// If the authentication was successful
	//
	// required: true
	Success bool `json:"success"`
}

// Authentication response
//
// swagger:response AuthResponse
type AuthResponseWrapper struct {
	// The response body
	//
	// in: body
	Body AuthResponse
}
