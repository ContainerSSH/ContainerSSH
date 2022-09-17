[![ContainerSSH - Launch Containers on Demand](https://containerssh.github.io/images/logo-for-embedding.svg)](https://containerssh.io/)

<!--suppress HtmlDeprecatedAttribute -->
<h1 align="center">ContainerSSH Health Check Library</h1>

[![Go Report Card](https://goreportcard.com/badge/github.com/containerssh/health?style=for-the-badge)](https://goreportcard.com/report/github.com/containerssh/health)

This is a health check service returning "ok" if all required ContainerSSH services are running.

<p align="center"><strong>⚠⚠⚠ Warning: This is a developer documentation. ⚠⚠⚠</strong><br />The user documentation for ContainerSSH is located at <a href="https://containerssh.io">containerssh.io</a>.</p>

## Using this service 

This library uses ContainerSSH's own [HTTP](https://github.com/containerssh/http) implementation to create an HTTP server that returns "ok" when all services are up.

You can instantiate this service as described in the [service library](https://github.com/containerssh/service) as follows:

```go
svc, err := health.New(
    health.Config{
        Enable: true
        ServerConfiguration: http.ServerConfiguration{
            Listen: "0.0.0.0:23074",
        },
    },
    logger)

if err != nil {
    // ...
}
```

You can change the `ok`/`not ok` status by calling `srv.ChangeStatus(bool)`, like so:

```go
srv.ChangeStatus(true)
```

## Health check client

This library also provides a built-in client for running health checks. This can be used as follows:

```go
client, err := health.NewClient(
    health.Config{
        Enable: true
        Client: http.ClientConfiguration{
            URL: "http://0.0.0.0:23074",
        },
    },
    logger)
)
if client.Run() {
    // Success
} else {
    // Failed
}
```
