package config

import "github.com/containerssh/libcontainerssh/metadata"

// Request is the request object passed from the client to the config server.
//
// swagger:model ConfigRequest
type Request struct {
	// Metadata is the metadata received from the authentication server.
	metadata.ConnectionAuthenticatedMetadata `json:",inline"`
}

// ResponseBody is the structure representing the JSON HTTP response.
//
// swagger:model ConfigResponseBody
type ResponseBody struct {
	// Metadata is the metadata received from the authentication server.
	metadata.ConnectionAuthenticatedMetadata `json:",inline"`

	// Config is the configuration structure to be passed back from the config server.
	//
	// required: true
	Config AppConfig `json:"config"`
}

// Response is the entire response from the config server
//
// swagger:response ConfigResponse
type Response struct {
	// Body is the configuration response body.
	//
	// in: body
	// required: true
	Body ResponseBody
}
