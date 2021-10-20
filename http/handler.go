package http

// RequestHandler is an interface containing a simple controller receiving a request and providing a response.
type RequestHandler interface {
	// OnRequest is a method receiving a request and is able to respond.
	OnRequest(request ServerRequest, response ServerResponse) error
}

// ServerRequest is a testdata structure providing decoding from the raw request.
type ServerRequest interface {
	// Decode decodes the raw request into the provided target from a JSON format. It provides an
	//        error if the decoding failed, which should be passed back through the request handler.
	Decode(target interface{}) error
}

// ServerResponse is a response structure that can be used by the RequestHandler to set the response details.
type ServerResponse interface {
	// SetStatus sets the HTTP status code
	SetStatus(statusCode uint16)
	// SetBody sets the object of the response which will be encoded as JSON.
	SetBody(interface{})
}
