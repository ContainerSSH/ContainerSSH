[![ContainerSSH - Launch Containers on Demand](https://containerssh.github.io/images/logo-for-embedding.svg)](https://containerssh.io/)

<!--suppress HtmlDeprecatedAttribute -->
<h1 align="center">ContainerSSH Security Library</h1>

[![Go Report Card](https://goreportcard.com/badge/github.com/containerssh/security?style=for-the-badge)](https://goreportcard.com/report/github.com/containerssh/security)
[![LGTM Alerts](https://img.shields.io/lgtm/alerts/github/ContainerSSH/security?style=for-the-badge)](https://lgtm.com/projects/g/ContainerSSH/security/)

This library provides a security overlay for the [sshserver](https://github.com/containerssh/sshserver) library.

<p align="center"><strong>⚠⚠⚠ Warning: This is a developer documentation. ⚠⚠⚠</strong><br />The user documentation for ContainerSSH is located at <a href="https://containerssh.io">containerssh.io</a>.</p>

## Using this library

This library is intended as a tie-in to an existing module and does not implement a full SSH backend. Instead, you can use the `New()` function to create a network connection handler with an appropriate backend:

```go
security, err := security.New(
    config,
    backend
)
```

The `backend` should implement the `sshserver.NetworkConnectionHandler` interface from the [sshserver](https://github.com/containerssh/sshserver) library. For the details of the configuration structure please see [config.go](config.go).
