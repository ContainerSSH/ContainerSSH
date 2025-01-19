<!--suppress HtmlDeprecatedAttribute -->
<h1 align="center">ContainerSSH Metrics Integration Library</h1>

This library integrates the [metrics service](https://github.com/containerssh/metrics) with the [sshserver library](https://github.com/containerssh/sshserver).

## Using this library

This library is intended as an overlay/proxy for a handler for the [sshserver library](https://github.com/containerssh/containerssh/tree/main/internal/sshserver) "handler". It can be injected transparently to collect the following metrics:

- `containerssh_ssh_connections`
- `containerssh_ssh_handshake_successful`
- `containerssh_ssh_handshake_failed`
- `containerssh_ssh_current_connections`
- `containerssh_auth_server_failures`
- `containerssh_auth_failures`
- `containerssh_auth_success`

The handler can be instantiated with the following method:

```go
handler, err := metricsintegration.New(
    config,
    metricsCollector,
    backend,
)
```

- `config` is a configuration structure from the [metrics library](https://github.com/containerssh/containerssh/tree/main/metrics). This is used to bypass the metrics integration backend if metrics are disabled.
- `metricsCollector` is the metrics collector from the [metrics library](https://github.com/containerssh/containerssh/tree/main/metrics).
- `backend` is an SSH server backend from the [sshserver library](https://github.com/containerssh/containerssh/tree/main/internal/sshserver).
