package message

// This message indicates that JSON encoding the request failed. This is usually a bug.
const EHTTPFailureEncodeFailed = "HTTP_CLIENT_ENCODE_FAILED"

// This message indicates a connection failure on the network level.
const EHTTPFailureConnectionFailed = "HTTP_CLIENT_CONNECTION_FAILED"

// This message indicates that decoding the JSON response has failed. The status code is set for this
// code.
const EHTTPFailureDecodeFailed = "HTTP_CLIENT_DECODE_FAILED"

// This message indicates that ContainerSSH is not following a HTTP redirect sent by the server. Use the allowRedirects
// option to allow following HTTP redirects.
const EHTTPClientRedirectsDisabled = "HTTP_CLIENT_REDIRECTS_DISABLED"

// This message indicates that a HTTP request is being sent from ContainerSSH
const MHTTPClientRequest = "HTTP_CLIENT_REQUEST"

// This message indicates that the server responded with a HTTP redirect.
const MHTTPClientRedirect = "HTTP_CLIENT_REDIRECT"

// This message indicates that ContainerSSH received a HTTP response from a server.
const MHTTPClientResponse = "HTTP_CLIENT_RESPONSE"

// The HTTP server failed to write the response.
const MHTTPServerResponseWriteFailed = "HTTP_SERVER_RESPONSE_WRITE_FAILED"

// The HTTP server failed to encode the response object. This is a bug, please report it.
const MHTTPServerEncodeFailed = "HTTP_SERVER_ENCODE_FAILED"
