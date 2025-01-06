[![ContainerSSH - Launch Containers on Demand](https://containerssh.github.io/images/logo-for-embedding.svg)](https://containerssh.github.io/)

<!--suppress HtmlDeprecatedAttribute -->
<h1 align="center">ContainerSSH SSH Server Audit Log Integration</h1>

[![Go Report Card](https://goreportcard.com/badge/github.com/containerssh/auditlogintegration?style=for-the-badge)](https://goreportcard.com/report/github.com/containerssh/auditlogintegration)
[![LGTM Alerts](https://img.shields.io/lgtm/alerts/github/ContainerSSH/auditlogintegration?style=for-the-badge)](https://lgtm.com/projects/g/ContainerSSH/auditlogintegration/)

This library provides an integration overlay for the [SSH server](https://github.com/containerssh/sshserver) and the [audit log library](https://github.com/containerssh/auditlog)

<p align="center"><strong>⚠⚠⚠ Warning: This is a developer documentation. ⚠⚠⚠</strong><br />The user documentation for ContainerSSH is located at <a href="https://containerssh.io">containerssh.io</a>.</p>

## Using this library

In order to use this library you will need two things:

1. An [audit logger from the auditlog library](https://github.com/containerssh/auditlog).
2. A [handler from the sshserver library](https://github.com/containerssh/sshserver) to act as a backend to this library.

You can then create the audit logging handler like this:

```go
handler := auditlogintegration.New(
    backend,
    auditLogger,
)
```

You can then pass this handler to the SSH server as [described in the readme](https://github.com/containerssh/sshserver):

```go
server, err := sshserver.New(
    cfg,
    handler,
    logger,
)
```
