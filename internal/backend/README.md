<!--suppress HtmlDeprecatedAttribute -->
<h1 align="center">ContainerSSH Backend Library</h1>

This library handles SSH requests and routes them to a container or other backends.

## Using this library

This library can be used in conjunction with the [sshserver](https://github.com/containerssh/sshserver) to route SSH connections to containers.

You can create a new backend handler like this:

```go
handler, err := backend.New(
    config,
    logger,
    loggerFactory,
    authBehavior,
)
```

This method accepts the following parameters:

`config`
: The `AppConfig` struct from the [configuration library](https://github.com/containerssh/configuration). This is needed because this library performs a call to the config server if configured to fetch a connection-specific information.

`logger`
: This variable is a logger from the [log library](https://github.com/containerssh/libcontainerssh/log).

`loggerFactory`
: This is a logger factory used by the backend to create a logger for the instantiated backends after fetching the connection-specific configuration.

`authBehavior`
: This variable can contain one of `sshserver.AuthResponseSuccess`, `sshserver.AuthResponseFailure`, or `sshserver.AuthResponseUnavailable` to indicate how the backend should react to authentication requests. Normally, this can be set to `sshserver.AuthResponseUnavailable` since the [auth integration library](https://github.com/containerssh/authintegration) will take care of the authentication.

The handler can be passed to the [sshserver](https://github.com/containerssh/sshserver) or to another overlay as a backend, for example [auth integration](https://github.com/containerssh/authintegration).
