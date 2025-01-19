[![ContainerSSH - Launch Containers on Demand](https://containerssh.github.io/images/logo-for-embedding.svg)](https://containerssh.github.io/)

<!--suppress HtmlDeprecatedAttribute -->
<h1 align="center">ContainerSSH Kubernetes Library</h1>

[![Go Report Card](https://goreportcard.com/badge/github.com/containerssh/kuberun?style=for-the-badge)](https://goreportcard.com/report/github.com/containerssh/kuberun)
[![LGTM Alerts](https://img.shields.io/lgtm/alerts/github/ContainerSSH/kuberun?style=for-the-badge)](https://lgtm.com/projects/g/ContainerSSH/kuberun/)

This library runs Kubernetes pods in integration with the [sshserver library](https://github.com/containerssh/containerssh/tree/main/internal/sshserver).

<p align="center"><strong>⚠⚠⚠ Warning: This is a developer documentation. ⚠⚠⚠</strong><br />The user documentation for ContainerSSH is located at <a href="https://containerssh.io">containerssh.io</a>.</p>

## How this library works

When a client successfully performs an SSH handshake this library creates a Pod in the specified Kubernetes cluster. This pod will run the command specified in `IdleCommand`. When the user opens a session channel this library runs an `exec` command against this container, allowing multiple parallel session channels to work on the same Pod.

## Using this library

As this library is designed to be used exclusively with the [sshserver library](https://github.com/containerssh/containerssh/tree/main/internal/sshserver) the API to use it is also very closely aligned. This backend doesn't implement a full SSH backend, instead it implements a network connection handler. This handler can be instantiated using the `kuberun.New()` method:

```go
handler, err := kuberun.New(
    client,
    connectionID,
    config,
    logger,
    backendRequestsCounter,
    backendFailuresCounter,
)
```

The parameters are as follows:

- `config` is a struct of the [`kuberun.Config` type](config.go).
- `connectionID` is an opaque ID for the connection.
- `client` is the `net.TCPAddr` of the client that connected.
- `logger` is the logger from the [log library](https://github.com/containerssh/containerssh/tree/main/log)
- `backendRequestsCounter` and `backendFailuresCounter` are counters from the [metrics library](https://github.com/containerssh/containerssh/tree/main/metrics)

Once the handler is created it will wait for a successful handshake:

```go
sshConnection, err := handler.OnHandshakeSuccess("username-here")
```

This will launch a pod. Conversely, the `handler.OnDisconnect()` will destroy the pod.

The `sshConnection` can be used to create session channels and launch programs as described in the [sshserver library](https://github.com/containerssh/containerssh/tree/main/internal/sshserver).

**Note:** This library does not perform authentication. Instead, it will always `sshserver.AuthResponseUnavailable`.
