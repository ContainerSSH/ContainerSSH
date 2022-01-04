package config

import (
	"github.com/containerssh/libcontainerssh/auth"
)

// Request is the request object passed from the client to the config server.
//
// swagger:model Request
type Request struct {
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
	Metadata *auth.ConnectionMetadata `json:"metadata"`
}

// ResponseBody is the structure representing the JSON HTTP response.
//
// swagger:model ResponseBody
type ResponseBody struct {
	// Config is the configuration structure to be passed back from the config server.
	//
	// required: true
	Config AppConfig `json:"config"`
}

// Response is the entire response from the config server
//
// swagger:response Response
type Response struct {
	// Body is the configuration response body.
	//
	// in: body
	// required: true
	Body ResponseBody
}
