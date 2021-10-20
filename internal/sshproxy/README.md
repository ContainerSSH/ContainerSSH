[![ContainerSSH - Launch Containers on Demand](https://containerssh.github.io/images/logo-for-embedding.svg)](https://containerssh.io/)

<!--suppress HtmlDeprecatedAttribute -->
<h1 align="center">ContainerSSH SSH Proxy Backend</h1>

[![Go Report Card](https://goreportcard.com/badge/github.com/containerssh/sshproxy?style=for-the-badge)](https://goreportcard.com/report/github.com/containerssh/sshproxy)
[![LGTM Alerts](https://img.shields.io/lgtm/alerts/github/ContainerSSH/sshproxy?style=for-the-badge)](https://lgtm.com/projects/g/ContainerSSH/sshproxy/)

This is the SSH proxy backend for ContainerSSH, which forwards connections to a backend SSH server.

<p align="center"><strong>⚠⚠⚠ Warning: This is a developer documentation. ⚠⚠⚠</strong><br />The user documentation for ContainerSSH is located at <a href="https://containerssh.io">containerssh.io</a>.</p>

## Using this library

This library implements a `NetworkConnectionHandler` from the [sshserver library](https://github.com/containerssh/sshserver). This can be embedded into a connection handler.

The network connection handler can be created with the `New()` method:

```go
var client net.TCPAddr
connectionID := "0123456789ABCDEF"
config := sshproxy.Config{
    //...
}
collector := metrics.New()
proxy, err := sshproxy.New(
    client,
    connectionID,
    config,
    logger,
    collector.MustCreateCounter("backend_requests", "", ""),
    collector.MustCreateCounter("backend_failures", "", ""),
)
if err != nil {
    // Handle error
}
```

The `logger` parameter is a logger from the [ContainerSSH logger library](https://github.com/containerssh/containerssh/log).