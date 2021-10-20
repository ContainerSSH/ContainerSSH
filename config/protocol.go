package config

// ConfigRequest is the request object passed from the client to the config server.
//
// swagger:model ConfigRequest
type ConfigRequest struct {
	// Username is the username passed during authentication.
	//
	// required: true
	Username string `json:"username"`
	// RemoteAddr is the IP address (IPv4 or IPv6) of the connecting user.
	//
	// required: true
	RemoteAddr string `json:"remoteAddr"`
	// ConnectionID is a unique opaque ID for the connection from the user.
	//
	// required: true
	ConnectionID string `json:"connectionId"`
	// SessionID is an alias for ConnectionID and will be removed in future versions.
	//
	// required: true
	SessionID string `json:"sessionId"`
	// Metadata is the metadata received from the authentication server.
	//
	// required: false
	Metadata map[string]string `json:"metadata"`
}

// ConfigResponseBody is the structure representing the JSON HTTP response.
//
// swagger:model ConfigResponseBody
type ConfigResponseBody struct {
	// Config is the configuration structure to be passed back from the config server.
	//
	// required: true
	Config AppConfig `json:"config"`
}

// ConfigResponse is the entire response from the config server
//
// swagger:response ConfigResponse
type ConfigResponse struct {
	// Body is the configuration response body.
	//
	// in: body
	// required: true
	Body ConfigResponseBody
}
