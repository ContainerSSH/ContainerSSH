<!--suppress HtmlDeprecatedAttribute -->
<h1 align="center">ContainerSSH Logging Library</h1>

This library provides internal logging for ContainerSSH. Its functionality is very similar to how syslog is structured.

## Basic concept

This is not the logger you would expect from the [go-log](https://github.com/go-log/log) library. This library combines the Go errors and the log messages into one. In other words, **the `Message` object can be used both as an error and a log message**.

The main `Message` structure has several properties: a unique error code, a user-safe error message and a detailed error message for the system administrator. Additionally, it also may contain several key-value pairs called labels to convey additional information.

## Creating a message

If you want to create a message you can use the following methods:

### `log.UserMessage()`

The `log.UserMessage` method creates a new `Message` with a user-facing message structure as follows:

```go
msg := log.UserMessage(
    "E_SOME_ERROR_CODE",
    "Dear user, an internal error happened.",
    "Details about the error (%s)",
    "Some string that will end up instead of %s."
)
```

- The first parameter is an error code that is unique so people can identify the error. This error code should be documented so users can find it easily.
- The second parameter is a string printed to the user. It does not support formatting characters.
- The third parameter is a string that can be logged for the system administrator. It can contain `fmt.Sprintf`-style formatting characters.
- All subsequent parameters are used to replace the formatting characters.

### `log.NewMessage()`

The `log.NewMessage()` method is a simplified version of `log.UserMessage()` without the user-facing message. The user-facing message will always be `Internal Error.`. The method signature is the following:

```go
msg := log.NewMessage(
    "E_SOME_ERROR_CODE",
    "Details about the error (%s)",
    "Some string that will end up instead of %s."
)
```

### `log.WrapUser()`

The `log.WrapUser()` method can be used to create a wrapped error with a user-facing message. It automatically appends the original error message to the administrator-facing message. The function signature is the following:

```go
msg := log.WrapUser(
    originalErr,
    "E_SOME_CODE",
    "Dear user, some error happened."
    "Dear admin, an error happened. %s" + 
        "The error message will be appended to this message."
    "This string will appear instead of %s in the admin-message."
)
```

### `log.Wrap()`

Like the `log.WrapUser()` method the `log.Wrap()` will skip the user-visible error message and otherwise be identical to `log.Wrap()`.

```go
msg := log.Wrap(
    originalErr,
    "E_SOME_CODE",
    "Dear admin, an error happened. %s" + 
        "The error message will be appended to this message."
    "This string will appear instead of %s in the admin-message."
)
```

### Adding labels to messages

Labels are useful for recording extra information with messages that can be indexed by the logging system. These labels may or may not be recorded by the logging backend. For example, the syslog output doesn't support recording labels due to size constraints. In other words, the message itself should contain enough information for an administrator to interpret the error. 

You can add labels to messages like this:

```go
msg.Label("labelName", "labelValue")
```

**Hint:** `Label()` calls can be chained.

## Using messages

As mentioned before, the `Message` interface implements the `error` interface, so these messages can simply be returned like a normal error would.

## Logging

This library also provides a `Logger` interface that can log all kinds of messages and errors, including the `Message` interface. It provides the following methods for logging:

- `logger.Debug(message ...interface{})`
- `logger.Info(message ...interface{})`
- `logger.Notice(message ...interface{})`
- `logger.Warning(message ...interface{})`
- `logger.Error(message ...interface{})`
- `logger.Critical(message ...interface{})`
- `logger.Alert(message ...interface{})`
- `logger.Emergency(message ...interface{})`

You are encouraged to pass a `Message` implementation to it. 

We also provide the following compatibility methods which log at the info level.

- `logger.Log(v ...interface{})`
- `logger.Logf(format string, v ...interface{})`

We provide a method to create a child logger that has a different minimum log level. Messages below this level will be discarded:

```go
newLogger := logger.WithLevel(log.LevelInfo)
```

We can also create a new logger copy with default labels added:

```go
newLogger := logger.WithLabel("label name", "label value")
```

Finally, the logger also supports calling the `Rotate()` and `Close()` methods. `Rotate()` instructs the output to close all handles and reopen them to facilitate rotating logs. `Close()` permanently closes the writer.

## Creating a logger

The `Logger` interface is intended for generic implementations. The default implementation can be created as follows:

```go
logger, err := log.NewLogger(config)
```

Alternatively, you can also use the `log.MustNewLogger` method to skip having to deal with the error. (It will `panic` if an error happens.)

If you need a factory you can use the `log.LoggerFactory` interface and the `log.NewLoggerFactory` to create a factory you can pass around. The `Make(config)` method will make a new logger when needed.

## Configuration

The configuration structure for the default logger implementation is contained in the `log.Config` structure.

### Configuring the output format

The most important configuration is where your logs will end up:

```go
log.Config{
    Output: log.OutputStdout,
}
```

The following options are possible:

- `log.OutputStdout` logs to the standard output or any other `io.Writer`
- `log.OutputFile` logs to a file on the local filesystem.
- `log.OutputSyslog` logs to a syslog server using a UNIX or UDP socket.

### Logging to stdout

If you set the `Output` option to `log.OutputStdout` the output will be written to the standard output in the format specified below (see "Changing the log format"). The destination can be overridden:

```go
log.Config {
    Output: log.OutputStdout,
    Stdout: someOtherWriter,
}
```

### Logging to a file

The file logger is configured as follows:

```go
log.Config {
    Output: log.OutputFile,
    File: "/var/log/containerssh.log",
}
```

You can call the `Rotate()` method on the logger to close and reopen the file. This allows for log rotation.

### Logging to syslog

The syslog logger writes to a syslog daemon. Typically, this is located on the `/dev/log` UNIX socket, but sending logs over UDP is also supported. TCP, encryption, and other advanced Syslog features are not supported, so adding a Syslog daemon on the local node is strongly recommended.

The configuration is the following:

```go
log.Config{
    Output: log.OutputSyslog,
    Facility: log.FacilityStringAuth, // See facilities list below
    Tag: "ProgramName", // Add program name here
    Pid: false, // Change to true to add the PID to the Tag
    Hostname: "" // Change to override host name 
}
```

The following facilities are supported:

- `log.FacilityStringKern`
- `log.FacilityStringUser`
- `log.FacilityStringMail`
- `log.FacilityStringDaemon`
- `log.FacilityStringAuth`
- `log.FacilityStringSyslog`
- `log.FacilityStringLPR`
- `log.FacilityStringNews`
- `log.FacilityStringUUCP`
- `log.FacilityStringCron`
- `log.FacilityStringAuthPriv`
- `log.FacilityStringFTP`
- `log.FacilityStringNTP`
- `log.FacilityStringLogAudit`
- `log.FacilityStringLogAlert`
- `log.FacilityStringClock`
- `log.FacilityStringLocal0`
- `log.FacilityStringLocal1`
- `log.FacilityStringLocal2`
- `log.FacilityStringLocal3`
- `log.FacilityStringLocal4`
- `log.FacilityStringLocal5`
- `log.FacilityStringLocal6`
- `log.FacilityStringLocal7`

### Changing the log format

We currently support two log formats: `text` and `ljson`. The format is applied for the stdout and file outputs and can be configured as follows:

```go
log.Config {
    Format: log.FormatText|log.FormatLJSON,
}
```

#### The `text` format

The text format is structured as follows:

```
TIMESTAMP[TAB]LEVEL[TAB]MODULE[TAB]MESSAGE[NEWLINE]
```

- `TIMESTAMP` is the timestamp of the message in RFC3339 format.
- `LEVEL` is the level of the message (`debug`, `info`, `notice`, `warning`, `error`, `critical`, `alert`, `emergency`)
- `MODULE` is the name of the module logged. May be empty.
- `MESSAGE` is the text message or structured data logged.

This format is recommended for human consumption only.

#### The `ljson` format

This format logs in a newline-delimited JSON format. Each message has the following format:

```json
{"timestamp": "TIMESTAMP", "level": "LEVEL", "module": "MODULE", "message": "MESSAGE", "details": "DETAILS"}
```

- `TIMESTAMP` is the timestamp of the message in RFC3339 format.
- `LEVEL` is the level of the message (`debug`, `info`, `notice`, `warning`, `error`, `critical`, `alert`, `emergency`)
- `MODULE` is the name of the module logged. May be absent if not sent.
- `MESSAGE` is the text message. May be absent if not set.
- `DETAILS` is a structured log message. May be absent if not set.

## Creating a logger for testing

You can create a logger for testing purposes that logs using the `t *testing.T` log facility:

```go
logger := log.NewTestLogger(t)
```

You may also want to enable GitHub Actions logging by creating the following method:

```go
func TestMain(m *testing.M) {
	log.RunTests(m)
}
```

## Generating message code files

This package also includes a utility to generate and update a [CODES.md](CODES.md) from a [codes.go](codes.go) file in your repository to create a documentation about message codes.

To use this utility your `codes.go` must contain message codes in individual const declarations with documentation, not as a `const` block. You can install this package as follows:

```
go get -u github.com/containerssh/containerssh/log/cmd/containerssh-generate-codes
``` 

You can then execute `containerssh-generate-codes` to generate `CODES.md`. Optionally, you can pass the filenames as parameters:

```
containerssh-generate-codes source.go DESTINATION.md
```

We recommend creating a `codes_doc.go` file with the following content:

```go
package yourpackage

//go:generate containerssh-generate-codes
```

This lets you generate `CODES.md` using `go generate`.
