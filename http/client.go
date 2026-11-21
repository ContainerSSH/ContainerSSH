package http

// Client is a simplified HTTP interface that ensures that a struct is transported to a remote endpoint
// properly encoded, and the response is decoded into the response struct.
type Client interface {
	// Request queries the specified path on the configured endpoint with the specified method.
	Request(
		method string,
		path string,
		requestBody interface{},
		responseBody interface{},
	) (statusCode int, err error)

	// RequestURL requests a URL irrespective of the endpoint configured on the client.
	RequestURL(
		method string,
		url string,
		requestBody interface{},
		responseBody interface{},
	) (statusCode int, err error)

	// Get queries the configured endpoint with the path providing the response in the responseBody structure. It
	// returns the HTTP status code and any potential errors.
	Get(
		path string,
		responseBody interface{},
	) (statusCode int, err error)

	// Post queries the configured endpoint with the path, sending the requestBody and providing the
	// response in the responseBody structure. It returns the HTTP status code and any potential errors.
	Post(
		path string,
		requestBody interface{},
		responseBody interface{},
	) (statusCode int, err error)

	// Put queries the configured endpoint with the path, sending the requestBody and providing the
	// response in the responseBody structure. It returns the HTTP status code and any potential errors.
	Put(
		path string,
		requestBody interface{},
		responseBody interface{},
	) (statusCode int, err error)

	// Patch queries the configured endpoint with the path, sending the requestBody and providing the
	// response in the responseBody structure. It returns the HTTP status code and any potential errors.
	Patch(
		path string,
		requestBody interface{},
		responseBody interface{},
	) (statusCode int, err error)

	// Delete queries the configured endpoint with the path, sending the requestBody and providing the
	// response in the responseBody structure. It returns the HTTP status code and any potential errors.
	Delete(
		path string,
		requestBody interface{},
		responseBody interface{},
	) (statusCode int, err error)
}
