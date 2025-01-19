<!--suppress HtmlDeprecatedAttribute -->
<h1 align="center">ContainerSSH Audit Logging Library</h1>

This is an audit logging library for [ContainerSSH](https://containerssh.github.io). Among others, it contains the encoder and decoder for the [ContainerSSH Audit Log Format](FORMAT.v1.md) written in Go. This readme will guide you through the process of using this library.

<p align="center"><strong>⚠⚠⚠ Warning: This is a developer documentation. ⚠⚠⚠</strong><br />The user documentation for ContainerSSH is located at <a href="https://containerssh.io">containerssh.io</a>.</p>

## Setting up a logging pipeline

This section will explain how to set up and use a logging pipeline. As a first step, you must create the logger. The easiest way to do that is to pass a config object. The `geoIPLookupProvider` is provided by the [GeoIP library](https://github.com/containerssh/geoip), while `logger` is a logger implementation from the [Log library](https://github.com/containerssh/containerssh/tree/main/log).

```go
auditLogger, err := auditlog.New(cfg, geoIPLookupProvider, logger)
```

The `cfg` variable must be of the type `auditlog.Config`. Here's an example configuration:

```go
config := auditlog.Config{
    Enable: true,
    Format:  "binary",
    Storage: "file",
    File: file.Config{
        Directory: "/tmp/auditlog",
    },
    Intercept: auditlog.InterceptConfig{
        Stdin:     true,
        Stdout:    true,
        Stderr:    true,
        Passwords: true,
    },
}
```

The `logger` variable must be an instance of `go.containerssh.io/containerssh/log/logger`. The easiest way to create the logger is as follows:

```go
logger := standard.New()
```

Alternatively, you can also create the audit logger using the following factory method:

```go
auditLogger := auditlog.NewLogger(
    intercept,
    encoder,
    storage,
    logger,
)
```

In this case `intercept` is of the type `InterceptConfig`, `encoder` is an instance of `codec.Encoder`, `storage` is an instance of `storage.WritableStorage`, and `logger` is the same logger as explained above. This allows you to create a custom pipeline.

You can also trigger a shutdown of the audit logger with the `Shutdown()` method. This method takes a context as an argument, allowing you to specify a grace time to let the audit logger finish background processes:

```go
auditLogger.Shutdown(
    context.WithTimeout(
        context.Background(),
        30 * time.Second,
    ),
)
```

**Note:** the logger is not guaranteed to shut down when the shutdown context expires. If there are still active connections being logged it will wait for those to finish and be written to a persistent storage before exiting. It may, however, cancel uploads to a remote storage.

### Writing to the pipeline

Once the audit logging pipeline is created you can then create your first entry for a new connection:

```go
connectionID := "0123456789ABCDEF"
connection, err := auditLogger.OnConnect(
    []byte("asdf"),
    net.TCPAddr{
        IP:   net.ParseIP("127.0.0.1"),
        Port: 2222,
        Zone: "",
    },
)
```

This will post a `connect` message to the audit log. The `connection` variable can then be used to send
subsequent connection-specific messages:

```go
connection.OnAuthPassword("foo", []byte("bar"))
connection.OnDisconnect()
```

The `OnNewChannelSuccess()` method also allows for the creation of a channel-specific audit logger that will log with the appropriate channel ID. 

## Retrieving and decoding messages

Once the messages are restored they can be retrieved by the same storage mechanism that was used to store them:

```go
storage, err := auditlog.NewStorage(config, logger)
if err != nil {
    log.Fatalf("%v", err)
}
// This only works if the storage type is not "none"
readableStorage := storage.(storage.ReadableStorage)
```

The readable storage will let you list audit log entries as well as fetch individual audit logs:

```go
logsChannel, errors := readableStorage.List()
for {
    finished := false
    select {
    case entry, ok := <-logsChannel:
        if !ok {
            finished = true
            break
        }
        // use entry.Name to reference a specific audit log
    case err, ok := <-errors:
        if !ok {
            finished = true
            break
        }
        if err != nil {
            // Handle err
        }
    }
    if finished {
        break
    }
}
```

Finally, you can fetch an individual audit log:

```go
reader, err := readableStorage.OpenReader(entry.Name)
if err != nil {
    // Handle error
}
```

The reader is now a standard `io.Reader`. 

## Decoding messages

Messages can be decoded with the reader as follows:

```go
// Set up the decoder
decoder := binary.NewDecoder()

// Decode messages
decodedMessageChannel, errorsChannel := decoder.Decode(reader)

for {
    finished := false
    select {
        // Fetch next message or error
        case msg, ok := <-decodedMessageChannel:
            if !ok {
                //Channel closed
                finished = true
                break
            } 
            //Handle messages
        case err := <-errorsChannel:
            if !ok {
                //Channel closed
                finished = true
                break
            } 
            // Handle error
    }
    if finished {
        break
    }
}
```

**Tip:** The `<-` signs are used with channels. They are used for async processing. If you are unfamiliar with them take a look at [Go by Example](https://gobyexample.com/channels).

**Note:** The Asciinema encoder doesn't have a decoder pair as the Asciinema format does not contain enough information to reconstruct the messages.

## Development

In order to successfully run the tests for this library you will need a working [Docker](https://www.docker.com/) or [Podman](https://podman.io/) setup to run `minio/minio` for the S3 upload.

### Manually encoding messages

If you need to encode messages by hand without a logger pipeline you can do so with an encoder implementation. This is normally not needed. We have two encoder implementations: the binary and the Asciinema encoders. You can use them like this:

```go
geoIPLookup, err := geoip.New(...)
// Handle error 
encoder := binary.NewEncoder(logger, geoIPLookup)
// Alternatively:
// encoder := asciinema.NewEncoder(logger)

// Initialize message channel
messageChannel := make(chan message.Message)
// Initialize storage backend
storage := YourNewStorage()

go func() {
    err := encoder.Encode(messageChannel, storage)
    if err != nil {
        log.Fatalf("failed to encode messages (%w)", err)        
    }
}()

messageChannel <- message.Message{
    //Fill in message details here
}
//make sure to close the message channel so the encoder knows no more messages will come.
close(messageChannel)
```

**Note:** The encoder will run until the message channel is closed, or a disconnect message is sent.

### Implementing an encoder and decoder

If you want to implement your own encoder for a custom format you can do so by implementing the `Encoder` interface in the [codec/abstract.go file](codec/abstract.go). Conversely, you can implement the `Decoder` interface to implement a decoder.

### Implementing a writable storage

In order to provide storages you must provide an `io.WriteCloser` with this added function:

```go
// Set metadata for the audit log. Can be called multiple times.
//
// startTime is the time when the connection started in unix timestamp
// sourceIp  is the IP address the user connected from
// username  is the username the user entered. The first time this method
//           is called the username will be nil, may be called subsequently
//           is the user authenticated.
SetMetadata(startTime int64, sourceIp string, username *string)
```

### Implementing a readable storage

In order to implement a readable storage you must implement the `ReadableStorage` interface in [storage/storage.go](storage/storage.go). You will need to implement the `OpenReader()` method to open a specific audit log and the `List()` method to list all available audit logs.

## Generating the format documentation

The format documentation is autogenerated using `go generate`.
