<!--suppress HtmlDeprecatedAttribute -->
<h1 align="center">ContainerSSH SSH Server Library</h1>

This library provides an overlay for the built-in go SSH server library that makes it easier to handle.

## Using this library

This library provides a friendlier way to handle SSH requests than with the built-in SSH library. the primary method of using this library is via the `Lifecycle` objects from the [service library](https://github.com/containerssh/service):

```go
// Create the server. See the description below for parameters.
server, err := sshserver.New(
    cfg,
    handler,
    logger,
)
if err != nil {
    // Handle configuration errors
    log.Fatalf("%v", err)
}
lifecycle := service.NewLifecycle(server)

defer func() {
    // The Run method will run the server and return when the server is shut down.
    // We are running this in a goroutine so the shutdown below can proceed after a minute.
    if err := lifecycle.Run(); err != nil {
        // Handle errors while running the server
    }
}()

time.Sleep(60 * time.Second)

// Shut down the server. Pass a context to indicate how long the server should wait
// for existing connections to finish. This function will return when the server
// has stopped. 
lifecycle.Stop(
    context.WithTimeout(
        context.Background(),
        30 * time.Second,
    ),
)
```

The `cfg` variable will be a `Config` structure as described in [config.go](config.go).

The `handler` variable must be an implementation of the [`Handler` interface described in handler.go](handler.go).

The `logger` variable needs to be an instance of the `Logger` interface from [github.com/containerssh/libcontainerssh/log](https://github.com/containerssh/libcontainerssh/log).

## Implementing a handler

The handler interface consists of multiple parts:

- The `Handler` is the main handler for the application providing several hooks for events. On new connections the `OnNetworkConnection` method is called, which must return a `NetworkConnectionHandler`
- The `NetworkConnectionHandler` is a handler for network connections before the SSH handshake is complete. It is called to perform authentication and return an `SSHConnectionHandler` when the authentication is successful.
- The `SSHConnectionHandler` is responsible for handling an individual SSH connection. Most importantly, it is responsible for providing a `SessionChannelHandler` when a new session channel is requested by the client.
- The `SessionChannelHandler` is responsible for an individual session channel (single program execution). It provides several hooks for setting up and running the program. Once the program execution is complete the channel is closed. You must, however, keep handling requests (e.g. window size change) during program execution.

A sample implementation can be found in the [test code](server_test.go) at the bottom of the file.

## About the `connectionID`

The `connectionID` parameter in the `OnNetworkConnection()` is a hexadecimal string uniquely identifying a connection. This ID can be used to track connection-related information across multiple subsystems (e.g. logs, audit logs, authentication and configuration requests, etc.)

## Testing a backend

This library contains a testing toolkit for running Linux commands against a backend. The primary resource for these tests will be the [conformance tests](test_conformance.go). To use these you must implement a set of factories that fulfill the following signature: `func(logger log.Logger) (sshserver.NetworkConnectionHandler, error)`.

These factories can then be used as follows:

```go
func TestConformance(t *testing.T) {
		var factories = map[string]func() (
            sshserver.NetworkConnectionHandler,
            error,
        ) {
    		"some-method": func(
                logger log.Logger,
            ) (sshserver.NetworkConnectionHandler, error) {
    			
    		},
    		"some-other-method": func(
                logger log.Logger,
            ) (sshserver.NetworkConnectionHandler, error) {
    			
    		},
    	}
    
    	sshserver.RunConformanceTests(t, factories)
}
```

The conformance tests will then attempt to execute a series of Linux interactions against the network connection handler and report features that have failed.

Alternatively, you can also use the components that make up the conformance tests separately.

### Creating a test user

The first step of using the test utilities is creating a test user. This can be done using the following calls:

```go
user := sshserver.NewTestUser(
    "test",
)
``` 

This use can then be configured with a random password using `RandomPassword()`, or set credentials using `SetPassword()` or `GenerateKey()`. These test users can then be passed to the test server or test client.

### Creating a test client

The test SSH client offers functionality over the Golang SSH client that helps with testing. The test client can be created as follows:

```go
sshclient := NewTestClient(
    serverIPAndPort,
    serverHostPrivateKey,
    user *TestUser,
    logger log.Logger,
)
```

Note that you must pass the servers private key in PEM format which will be used to extract the public key for validation. This makes the client unsuitable for purposes other than testing.

The test client can then be used to interact with the server. The client is described in [test_client.go](test_client.go) and functions can be discovered using code completion in the IDE.

### Creating a test server

We also provide a simplified test server that you can plug in any backend:

```go
srv := NewTestServer(
    handler,
    logger,
)
```

You can then start the server in the background and subsequently stop it:

```go
srv.Start()
defer srv.Stop(10 * time.Second)
```

### Creating a simple authenticating handler

We also provide an authentication handler that can be used to authenticate using the test users:

```go
handler := NewTestAuthenticationHandler(
    handler,
    user1,
    user2,
    user3,
)
```
