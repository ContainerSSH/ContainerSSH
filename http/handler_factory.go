package http

import (
	goHttp "net/http"

    "go.containerssh.io/libcontainerssh/log"
)

// NewServerHandler creates a new simplified HTTP handler that decodes JSON requests and encodes JSON responses.
func NewServerHandler(
	requestHandler RequestHandler,
	logger log.Logger,
) goHttp.Handler {
	if requestHandler == nil {
		panic("BUG: no requestHandler provided to http.NewServerHandler")
	}
	if logger == nil {
		panic("BUG: no logger provided to http.NewServerHandler")
	}
	return &handler{
		requestHandler:            requestHandler,
		logger:                    logger,
		defaultResponseMarshaller: &jsonMarshaller{},
		defaultResponseType:       "application/json",
		responseMarshallers: []responseMarshaller{
			&jsonMarshaller{},
		},
	}
}

// NewServerHandlerNegotiate creates a simplified HTTP handler that supports content negotiation for responses.
//goland:noinspection GoUnusedExportedFunction
func NewServerHandlerNegotiate(
	requestHandler RequestHandler,
	logger log.Logger,
) goHttp.Handler {
	if requestHandler == nil {
		panic("BUG: no requestHandler provided to http.NewServerHandler")
	}
	if logger == nil {
		panic("BUG: no logger provided to http.NewServerHandler")
	}
	return &handler{
		requestHandler:            requestHandler,
		logger:                    logger,
		defaultResponseMarshaller: &jsonMarshaller{},
		defaultResponseType:       "application/json",
		responseMarshallers: []responseMarshaller{
			&jsonMarshaller{},
			&textMarshaller{},
		},
	}
}
