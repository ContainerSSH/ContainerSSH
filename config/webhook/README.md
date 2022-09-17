<!--suppress HtmlDeprecatedAttribute -->
<h1 align="center">ContainerSSH Configuration Library</h1>

This library provides configuration client and server components for dynamic SSH configuration.

## Creating a configuration server

The main use case of this library will be creating a configuration server to match the current release of ContainerSSH in Go.

First, you need to fetch this library as a dependency using [go modules](https://blog.golang.org/using-go-modules):

```bash
go get github.com/containerssh/configuration
```

Next, you will have to write an implementation for the following interface:

```go
type ConfigRequestHandler interface {
	OnConfig(request configuration.ConfigRequest) (configuration.AppConfig, error)
}
```

The best way to do this is creating a struct and adding a method with a receiver:

```go
type myConfigReqHandler struct {
}

func (m *myConfigReqHandler) OnConfig(
    request configuration.ConfigRequest,
) (config configuration.AppConfig, err error) {
    // We recommend using an IDE to discover the possible options here.
    if request.Username == "foo" {
        config.Docker.Config.ContainerConfig.Image = "yourcompany/yourimage"
    }
    return config, err
}
```

**Warning!** Your `OnConfig` method should *only* return an error if it can genuinely not serve the request. This should not be used as a means to reject users. This should be done using the [authentication server](https://github.com/containerssh/auth). If you return an error ContainerSSH will retry the request several times in an attempt to work around network failures.

Once you have your handler implemented you must decide which method you want to use for integration.

### The full server method

This method is useful if you don't want to run anything else on the webserver, only the config endpoint. You can create a new server like this:

```go
srv, err := configuration.NewServer(
	config.HTTPServerConfiguration{
        Listen: "0.0.0.0:8080",
    },
	&myConfigReqHandler{},
	logger,
)
```

The `logger` parameter is a logger from the [ContainerSSH log library](https://github.com/containerssh/libcontainerssh/log).

Once you have the server you can start it using the [service library](https://github.com/containerssh/service):

```go
lifecycle := service.NewLifecycle(srv)
err := lifecycle.Run()
```

This will run your server in an endless fashion. However, for a well-behaved server you should also implement signal handling:

```go
srv, err := webhook.NewServer(
    http.ServerConfiguration{
        Listen: "0.0.0.0:8080",
    },
    &myConfigReqHandler{},
    logger,
)
if err != nil {
    // Handle error
}

lifecycle := service.NewLifecycle(srv)

go func() {
    //Ignore error, handled later.
    _ = lifecycle.Run()
}()

signals := make(chan os.Signal, 1)
signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
go func() {
    if _, ok := <-signals; ok {
        // ok means the channel wasn't closed, let's trigger a shutdown.
        lifecycle.Shutdown(
            context.WithTimeout(
                context.Background(),
                20 * time.Second,
            )
        )
    }
}()
// Wait for the service to terminate.
lastError := lifecycle.Wait()
// We are already shutting down, ignore further signals
signal.Ignore(syscall.SIGINT, syscall.SIGTERM)
// close signals channel so the signal handler gets terminated
close(signals)

if err != nil {
    // Exit with a non-zero signal
    fmt.Fprintf(
        os.Stderr,
        "an error happened while running the server (%v)",
        err,
    )
    os.Exit(1)
}
os.Exit(0)
```

**Note:** We recommend securing client-server communication with certificates. The details about securing your HTTP requests are documented in the [HTTP library](https://github.com/containerssh/http).

### Integrating with an existing HTTP server

Use this method if you want to integrate your handler with an existing Go HTTP server. This is rather simple:

```go
handler, err := configuration.NewHandler(&myConfigReqHandler{}, logger)
```

You can now use the `handler` variable as a handler for the [`http` package](https://golang.org/pkg/net/http/) or a MUX like [gorilla/mux](https://github.com/gorilla/mux).

## Using the config client

This library also contains the components to call the configuration server in a simplified fashion. To create a client simply call the following method:

```go
client, err := configuration.NewClient(
	configuration.ClientConfig{
        http.ClientConfiguration{
            URL: "http://your-server/config-endpoint/"
        }
    },
	logger,
    metricsCollector,
)
```

The `logger` is a logger from the [log library](https://github.com/containerssh/libcontainerssh/log), the `metricsCollector` is supplied by the [metrics library](https://github.com/containerssh/metrics). 

You can now use the `client` variable to fetch the configuration specific to a connecting client:

```go
connectionID := "0123456789ABCDEF"
appConfig, err := client.Get(
    ctx,
    "my-name-is-trinity",
    net.TCPAddr{
        IP: net.ParseIP("127.0.0.1"),
        Port: 2222,
    },
    connectionID,
) (AppConfig, error)
```

Now you have the client-specific configuration in `appConfig`.

**Note:** We recommend securing client-server communication with certificates. The details about securing your HTTP requests are documented in the [HTTP library](https://github.com/containerssh/http).

## Loading the configuration from a file

This library also provides simplified methods for reading the configuration from an `io.Reader` and writing it to an `io.Writer`.

```go
file, err := os.Open("file.yaml")
// ...
loader, err := configuration.NewReaderLoader(
	file,
    logger,
    configuration.FormatYAML,
)
// Read global config
appConfig := &configuration.AppConfig{}
err := loader.Load(ctx, appConfig)
// Read connection-specific config:
err := loader.LoadConnection(
    ctx,
    "my-name-is-trinity",
    net.TCPAddr{
        IP: net.ParseIP("127.0.0.1"),
        Port: 2222,
    },
    connectionID,
    appConfig,
)
```

As you can see these loaders are designed to be chained together. For example, you could add an HTTP loader after the file loader:

```go
httpLoader, err := configuration.NewHTTPLoader(clientConfig, logger)
```

This HTTP loader calls the HTTP client described above.

Conversely, you can write the configuration to a YAML format:

```go
saver, err := configuration.NewWriterSaver(
    os.Stdout,
    logger,
    configuration.FormatYAML,
)
err := saver.Save(appConfig)
```
