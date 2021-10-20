[![ContainerSSH - Launch Containers on Demand](https://containerssh.github.io/images/logo-for-embedding.svg)](https://containerssh.github.io/)

<!--suppress HtmlDeprecatedAttribute -->
<h1 align="center">ContainerSSH Service Library</h1>

[![Go Report Card](https://goreportcard.com/badge/github.com/containerssh/service?style=for-the-badge)](https://goreportcard.com/report/github.com/containerssh/service)
[![LGTM Alerts](https://img.shields.io/lgtm/alerts/github/ContainerSSH/service?style=for-the-badge)](https://lgtm.com/projects/g/ContainerSSH/service/)

This library provides a common way to manage multiple independent services in a single binary.

<p align="center"><strong>⚠⚠⚠ Warning: This is a developer documentation. ⚠⚠⚠</strong><br />The user documentation for ContainerSSH is located at <a href="https://containerssh.io">containerssh.io</a>.</p>

## Creating a service

In order to create a service you must implement the [`Service` interface](service.go):

```go
type Service interface {
	// String returns the name of the service
	String() string

	RunWithLifecycle(lifecycle Lifecycle) error
}
```

The Run function gets passed a [`Lifecycle`](lifecycle.go) object. It must call the appropriate lifecycle hooks:

```go
func (s *myService) RunWithLifecycle(lifecycle Lifecycle) error {
    //Do initialization here
    lifecycle.Running()
    for {
        // Do something
        if err != nil {
            return err
        }
        if lifecycle.ShouldStop() {
            shutdownContext := lifecycle.Stopping()
            // Handle graceful shutdown.
            // If shutdownContext expires, shut down immediately.
            // Then exit out of the loop.
            break
        }
    }
    return nil
}
```

For advanced use cases you can replace the `lifecycle.ShouldStop()` call with fetching the context directly using `lifecycle.Context()`. You can then use the context in a `select` statement.

**Warning!** Do not call `RunWithLifecycle()` on the service directly. Instead, always call `Run()` on the lifecycle to enable accurate state tracking and error handling. 

## Creating a lifecycle

In order to run a service you need to create a `Lifecycle` object. Since `Lifecycle` is an interface you can implement it yourself, or you can use the default implementation:

```
lifecycle := service.NewLifecycle(service)
```

The `service` parameter should be the associated service. The lifecycle can be used to add hooks to the service. Calling these functions multiple times is supported, but the call order of hook functions is not guaranteed.

```go
lifecycle.OnStateChange(func(s service.Service, l service.Lifecycle, newState service.State) {
    // do something
})
lifecycle.OnStarting(func(s service.Service, l service.Lifecycle) {
    // do something
})
lifecycle.OnRunning(func(s service.Service, l service.Lifecycle) {
    // do something
})
lifecycle.OnStopping(func(s service.Service, l service.Lifecycle, shutdownContext context.Context) {
    // do something
})
lifecycle.OnStopped(func(s service.Service, l service.Lifecycle) {
    // do something
})
lifecycle.OnCrashed(func(s service.Service, l service.Lifecycle, err error) {
    // do something
})
```

These hook functions can also be chained:

```go
lifecycle.OnStarting(myHandler).OnRunning(myHandler)
```

You can now use the Lifecycle to run the service:

```go
err := lifecycle.Run()
```

**Warning!** Do not call `RunWithLifecycle()` on the service directly. Instead, always call `Run()` on the lifecycle to enable accurate state tracking and error handling.

## Using the service pool

One of the advanced components in this library is the `Pool` object. It provides an overlay for managing multiple services in parallel, and it implements the `Service` interface itself. In other words, it can be nested.

First, let's create a pool: 

```go
pool := service.NewPool(
    service.NewLifecycleFactory(),
    logger,
)
```

The `logger` variable is a logger from [the log package](https://github.com/containerssh/containerssh/log). You can then add subservices to the pool. When adding a service the pool will return the lifecycle object you can use to add hooks. The hook functions can be chained for easier configuration:

```go
_ = pool.
    Add(myService1).
    OnRunning(func (s Service, l Lifecycle) {
        log.Printf("%s is now %s", s.String(), l.State())
    })
```

Once the services are added the pool can be launched:

```go
lifecycle := service.NewLifecycle(pool)
go func() {
    err := lifecycle.Run()
    // Handle errors here
}
lifecycle.Shutdown(context.Background())
```

Ideally, the pool can be used to handle Ctrl+C and SIGTERM events:

```go
signals := make(chan os.Signal, 1)
signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
go func() {
    if _, ok := <-signals; ok {
        // ok means the channel wasn't closed
        lifecycle.Shutdown(
            context.WithTimeout(
                context.Background(),
                20 * time.Second,
            )
        )
    }
}()
// Wait for the pool to terminate.
lifecycle.Wait()
// We are already shutting down, ignore further signals
signal.Ignore(syscall.SIGINT, syscall.SIGTERM)
// close signals channel so the signal handler gets terminated
close(signals)
```
