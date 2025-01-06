<!--suppress HtmlDeprecatedAttribute -->
<h1 align="center">ContainerSSH HTTP Library</h1>

This library provides a common layer for HTTP clients and servers in use by ContainerSSH.

## Using this library

This library provides a much simplified API for both the HTTP client and server.

### Using the client

The client library takes a request object that [can be marshalled into JSON format](https://gobyexample.com/json) and sends it to the server. It then fills a response object with the response received from the server. In code:

```go
// Logger is from the github.com/containerssh/libcontainerssh/log package
logger := standard.New()
clientConfig := http.ClientConfiguration{
    URL:        "http://127.0.0.1:8080/",
    Timeout:    2 * time.Second,
    // You can add TLS configuration here:
    CaCert:     "Add expected CA certificate(s) here.",
                // CaCert is required for https:// URLs on Windows due to golang#16736
    // Optionally, for client authentication:
    ClientCert: "Client certificate in PEM format or file name",
    ClientKey:  "Client key in PEM format or file name",
    // Optional: switch to www-urlencoded request body
    RequestEncoding: http.RequestEncodingWWWURLEncoded,
}
client, err := http.NewClient(clientConfig, logger)
if err != nil {
    // Handle validation error
}

request := yourRequestStruct{}
response := yourResponseStruct{}

responseStatus, err := client.Post(
    context.TODO(),
    "/relative/path/from/base/url",
    &request,
    &response,
)
if err != nil {
    // Handle connection error
    clientError := &http.ClientError{}
    if errors.As(err, clientError) {
        // Grab additional information here
    } else {
    	// This should never happen
    }
}

if responseStatus > 399 {
    // Handle error
}
```

The `logger` parameter is a logger from the [github.com/containerssh/libcontainerssh/log](https://github.com/containerssh/libcontainerssh/log) package.

### Using the server

The server consist of two parts: the HTTP server and the handler. The HTTP server can be used as follows:

```go
server, err := http.NewServer(
    "service name",
    http.ServerConfiguration{
        Listen:       "127.0.0.1:8080",
        // You can also add TLS configuration
        // and certificates here:
        Key:          "PEM-encoded key or file name to cert here.",
        Cert:         "PEM-encoded certificate chain or file name here",
        // Authenticate clients with certificates:
        ClientCACert: "PEM-encoded client CA certificate or file name here",
    },
    handler,
    logger,
    func (url string) {
        fmt.Printf("Server is now ready at %s", url)
    }
)
if err != nil {
    // Handle configuration error
}
// Lifecycle from the github.com/containerssh/service package
lifecycle := service.NewLifecycle(server)
go func() {
    if err := lifecycle.Run(); err != nil {
        // Handle error
    }
}()
// Do something else, then shut down the server.
// You can pass a context for the shutdown deadline.
lifecycle.Shutdown(context.Background())
```

Like before, the `logger` parameter is a logger from the [github.com/containerssh/libcontainerssh/log](https://github.com/containerssh/libcontainerssh/log) package. The `handler` is a regular [go HTTP handler](https://golang.org/pkg/net/http/#Handler) that satisfies this interface:

```go
type Handler interface {
    ServeHTTP(http.ResponseWriter, *http.Request)
}
```

The lifecycle object is one from the [ContainerSSH service package](https://github.com/containerssh/service).

## Using a simplified handler

This package also provides a simplified handler that helps with encoding and decoding JSON messages. It can be created as follows:

```go
handler := http.NewServerHandler(yourController, logger)
```

The `yourController` variable then only needs to implement the following interface:

```go
type RequestHandler interface {
	OnRequest(request ServerRequest, response ServerResponse) error
}
```

For example:

```go
type MyRequest struct {
    Message string `json:"message"`
}

type MyResponse struct {
    Message string `json:"message"`
}

type myController struct {
}

func (c *myController) OnRequest(request http.ServerRequest, response http.ServerResponse) error {
    req := MyRequest{}
	if err := request.Decode(&req); err != nil {
		return err
	}
	if req.Message == "Hi" {
		response.SetBody(&MyResponse{
			Message: "Hello world!",
		})
	} else {
        response.SetStatus(400)
		response.SetBody(&MyResponse{
			Message: "Be nice and greet me!",
		})
	}
	return nil
}
```

In other words, the `ServerRequest` object gives you the ability to decode the request into a struct of your choice. The `ServerResponse`, conversely, encodes a struct into the response body and provides the ability to enter a status code.

## Content negotiation

If you wish to perform content negotiation on the server side, this library now supports switching between text and JSON output. This can be invoked using the `NewServerHandlerNegotiate` method instead of `NewServerHandler`. This handler will attempt to switch based on the `Accept` header sent by the client. You can marshal objects to text by implementing the following interface:

```go
type TextMarshallable interface {
	MarshalText() string
}
```

## Using multiple handlers

This is a very simple handler example. You can use utility like [gorilla/mux](https://github.com/gorilla/mux) as an intermediate handler between the simplified handler and the server itself.
