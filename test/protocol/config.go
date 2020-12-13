package protocol

import (
	"github.com/containerssh/configuration"
)

//
// swagger:model ConfigRequest
type ConfigRequest struct {
	Username  string `json:"username"`
	SessionId string `json:"sessionId"`
}

// swagger:model ConfigResponseBody
type ConfigResponse struct {
	Config configuration.AppConfig `json:"config"`
}

// The configuration response object
//
// swagger:response ConfigResponse
type ConfigResponseWrapper struct {
	// The configuration response body
	//
	// in: body
	// required: true
	Body ConfigResponse
}
