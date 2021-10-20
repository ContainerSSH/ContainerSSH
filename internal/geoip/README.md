<!--suppress HtmlDeprecatedAttribute -->
<h1 align="center">ContainerSSH IP to Country Code Library</h1>

This library provides IP to Country Code lookup services for ContainerSSH.

## Using this library

This library needs a configuration structure described in [config.go in the public package](../../config/geoip/config.go). This configuration structure can be passed to the `geoip.New()` method:

```go
provider, err := geoip.New(geoip.Config{
    // Can be "dummy" or "maxmind".
    Provider: "maxmind",
    // MMDB2 file for the MaxMind provider.
    GeoIP2File: "/path/to/maxmind/file.mmdb2",
})
if err != nil {
    // handle error
}
```

The GeoIP lookup can be performed using the `Lookup()` method:

```go
countryCode := provider.Lookup("127.0.0.1")
```

The `countryCode` field will contain the value of `XX` if the lookup failed.

## Implementing a lookup provider

A custom provider can be written by implementing the following interface:

```go
type LookupProvider interface {
	Lookup(remoteAddr net.IP) (countryCode string)
}
```

Once implemented you will need to add the necessary configuration options to [config.go](config.go) and add a factory method to [factory.go](factory.go).
