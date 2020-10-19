<h1>Metrics {{ since("0.3.0") }}</h1>

ContainerSSH contains a [Prometheus](https://prometheus.io/)-compatible metrics server which can be enabled using the following configuration:

```yaml
metrics:
  enable: true # Defaults to false
  listen: "0.0.0.0:9100" # Set the listen address here
  path: "/metrics" # Defaults to /metrics
```
