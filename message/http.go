package message

// EHTTPListenFailed indicates that ContainerSSH failed to listen on the specified port.
const EHTTPListenFailed = "HTTP_LISTEN_FAILED"

// EHTTPFailureEncodeFailed indicates that JSON encoding the request failed. This is usually a bug.
const EHTTPFailureEncodeFailed = "HTTP_CLIENT_ENCODE_FAILED"

// EHTTPFailureConnectionFailed indicates that a connection failure on the network level.
const EHTTPFailureConnectionFailed = "HTTP_CLIENT_CONNECTION_FAILED"

// EHTTPFailureDecodeFailed indicates that decoding the JSON response has failed. This is usually a bug in the webhook
// server.
const EHTTPFailureDecodeFailed = "HTTP_CLIENT_DECODE_FAILED"

// EHTTPClientRedirectsDisabled indicates that ContainerSSH is not following an HTTP redirect sent by the server. Use the
// allowRedirects option to allow following HTTP redirects or enter the config URL directly.
const EHTTPClientRedirectsDisabled = "HTTP_CLIENT_REDIRECTS_DISABLED"

// MHTTPClientRequest indicates that an HTTP request is being sent from ContainerSSH.
const MHTTPClientRequest = "HTTP_CLIENT_REQUEST"

// MHTTPClientRedirect indicates that the server responded with an HTTP redirect.
const MHTTPClientRedirect = "HTTP_CLIENT_REDIRECT"

// MHTTPClientResponse indicates that ContainerSSH received an HTTP response from a server.
const MHTTPClientResponse = "HTTP_CLIENT_RESPONSE"

// MHTTPServerResponseWriteFailed indicates that the HTTP server failed to write the response.
const MHTTPServerResponseWriteFailed = "HTTP_SERVER_RESPONSE_WRITE_FAILED"

// MHTTPServerEncodeFailed indicates that the HTTP server failed to encode the response object. This can happen
// if the incorrect response object was returned in a webhook.
const MHTTPServerEncodeFailed = "HTTP_SERVER_ENCODE_FAILED"
