# ContainerSSH Audit Log Decoder Library

This is a decoder library for the [ContainerSSH Audit Log Format](https://containerssh.github.io/audit/format/) written in Go. In order to use it you will need depend on `github.com/containerssh/containerssh/audit/format`.

You can read a stored and gzipped audit log file and decode it as follows:

```go
messages, errors, done := format.Decode(auditLogFileReader)
```

The `messages` variable will be a channel where messages are sent in-order and `errors` will be a channel where decoding errors are sent. The messages are structs of `DecodedMessage` detailed in [decode.go](decode.go).

The `payload` field in the message will contain a specific struct based on the `type` field detailing the message. You can find details about the mapping in the [log format description](https://containerssh.github.io/audit/format/).