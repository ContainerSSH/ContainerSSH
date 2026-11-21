<!--suppress HtmlDeprecatedAttribute -->
<h1 align="center">ContainerSSH Metrics Library</h1>

This library provides centralized metrics collection across modules. It also provides a [Prometheus](https://prometheus.io/) / [OpenMetrics](https://openmetrics.io/) compatible HTTP server exposing the collected metrics.

## Collecting metrics

The core component of the metrics is the `metrics.Collector` interface. You can create a new instance of this interface by calling `metrics.New()` with a GeoIP lookup provider from the [geoip library](https://github.com/containerssh/containerssh/tree/main/internal/geoip) as a parameter. You can then dynamically create metrics:

```go
m := metrics.New(geoip)
testCounter, err := m.CreateCounter(
    "test", // Name of the metric
    "MB", // Unit of the metric
    "This is a test", // Help text of the metric
)
```

You can then increment the counter:

```go
testCounter.Increment()
testCounter.IncrementBy(5)
```

Alternatively, you can also create a CounterGeo to make a label automatically based on GeoIP lookup:

```go
testCounter, err := m.CreateCounterGeo(
    "test", // Name of the metric
    "MB", // Unit of the metric
    "This is a test", // Help text of the metric
)
testCounter.Increment(net.ParseIP("127.0.0.1"))
```

If you need a metric that can be decremented or set directly you can use the `Gauge` type instead. Each `Create*` method also has a counterpart named `MustCreate*`, which panics instead of returning an error.

### Custom labels

Each of the metric methods allow adding extra labels:

```go
testCounter.Increment(
    net.ParseIP("127.0.0.1"),
    metrics.Label("foo", "bar"),
    metrics.Label("somelabel","somevalue")
)
```

The following rules apply and will cause a `panic` if violated:

- Label names and values cannot be empty.
- The `country` label name is reserved for GeoIP usage.

The metrics also have a `WithLabels()` method that allow for creating a copy of a metric already primed with a set of labels. This can be used when passing metrics to other modules that need to be scoped.

## Using the metrics server

The metrics server exposes the collected metrics on an HTTP webserver in the Prometheus / OpenMetrics format. It requires the [service library](https://github.com/containerssh/containerssh/tree/main/service) and a logger from the [log library](https://github.com/containerssh/containerssh/tree/main/log) to work properly:

```go
server := metrics.NewServer(
    metrics.Config{
        ServerConfiguration: http.ServerConfiguration{
            Listen:       "127.0.0.1:8080",
        },
        Enable:              true,
        Path:                "/metrics",
    },
    metricsCollector,
    logger,
)

lifecycle := service.NewLifecycle(server)
go func() {
    if err := lifecycle.Run(); err != nil {
        // Handle crash
    } 	
}()

//Later:
lifecycle.Stop(context.Background())
```

Alternatively, you can skip the full HTTP server and request a handler that you can embed in any Go HTTP server:

```go
handler := metrics.NewHandler(
    "/metrics",
    metricsCollector
)
http.ListenAndServe("0.0.0.0:8080", handler)
```

