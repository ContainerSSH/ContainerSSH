<h1>Metrics {{ since("0.3.0") }}</h1>

ContainerSSH contains a [Prometheus](https://prometheus.io/)-compatible metrics server which can be enabled using the following configuration:

```yaml
metrics:
  enable: true # Defaults to false
  listen: "0.0.0.0:9100" # Set the listen address here
  path: "/metrics" # Defaults to /metrics
```

You can configure Prometheus to grab the following metrics:

- `containerssh_auth_server_failures`: Number failed requests to the authentication server since start.
- `containerssh_auth_success`: Number of successful authentications since start. Contains labels for `authtype` (`password` or `pubkey`) and `country` (see below).
- `containerssh_auth_failures`: Number of failed authentications since start. Contains labels for `authtype` (`password` or `pubkey`) and `country` (see below).
- `containerssh_config_server_failures`: Number of failed requests to the configuration server since start.
- `containerssh_ssh_connections`: Number of SSH connections since since start. Contains a label for `country` (see below).
- `containerssh_ssh_handshake_successful`: Number of successful SSH handshakes since start. Contains a label for `country` (see below).
- `containerssh_ssh_handshake_failed`: Number of failed SSH handshakes since start. Contains a label for `country` (see below).
- `containerssh_ssh_current_connections`: Number of currently open SSH connections. Contains a label for `country` (see below).

## Country identification

Country identification works using [GeoIP2 or GeoLite2 from MaxMind](https://www.maxmind.com/en/geoip2-services-and-databases). This database needs to be provided to ContainerSSH externally due to licensing concerns. Currently we do not support other w

The default path for the GeoIP database is `/var/lib/GeoIP/GeoIP2-Country.mmdb`, but you can change that using the following configuration snippet:

```yaml
geoip:
  maxmind-geoip2-file: '/var/lib/GeoIP/GeoIP2-Country.mmdb'
```
