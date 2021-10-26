[![ContainerSSH - Launch Containers on Demand](https://containerssh.github.io/images/logo-for-embedding.svg)](https://containerssh.github.io/)

<!--suppress HtmlDeprecatedAttribute -->
<h1 align="center">libcontainerssh</h1>

This library is the core library of ContainerSSH and simultaneously serves as a library to integrate with ContainerSSH. This readme outlines the basics of using this library, for more detailed documentation please head to [containerssh.io](https://containerssh.io).

## Embedding ContainerSSH

You can fully embed ContainerSSH into your own application. First, you will need to create the configuration structure:

```go
cfg := config.AppConfig{}
```

You can then populate this config with your options and create a ContainerSSH instance like this:

```go
 pool, lifecycle, err := containerssh.New(cfg, loggerFactory)
 if err != nil {
     return err
 }
```

You will receive a service pool and a lifecycle as a response. You can use these to start the service pool of ContainerSSH. This will block execution until ContainerSSH stops.

```go
err := lifecycle.Run()
```

This will run ContainerSSH in the current Goroutine. You can also use the lifecycle to add hooks to lifecycle states of ContainerSSH. You must do this *before* you call `Run()`. For example:

```go
lifecycle.OnStarting(
    func(s service.Service, l service.Lifecycle) {
        print("ContainerSSH is starting...")
    },
)
```

You can also have ContainerSSH stop gracefully by using the `Stop()` function on the lifecycle. This takes a context as an argument, which is taken as a timeout for the graceful shutdown.

## Building an authentication webhook server

## Building a configuration webhook server

## Building a combined configuration-authentication webhook server

## Reading audit logs
