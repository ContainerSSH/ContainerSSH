<h1>Logging {{ since("0.2.2") }}</h1>

ContainerSSH comes with configurable logging facilities. At this time only JSON logging is supported, but the log level can be configured.

The configuration can be done from the config file:

```
log:
 level: "warning"
```

!!! tip
    You can configure the log level on a per-user basis using the [configuration server](configserver.md).

The supported levels are in accordance with the Syslog standard:

- `debug`
- `info`
- `notice`
- `warning`
- `error`
- `crit`
- `alert`
- `emerg`

## The JSON log format

The JSON log format outputs one line to the output per message. The message format is:

```json
{
  "timestamp":"Timestamp in RFC3339 format",
  "level":"the log level",
  "message":"the message (optional)",
  "details": {
    "the detail object if any (optional)"
  }
}
```

!!! note
    The JSON logger writes to the standard output regardless of log level.

!!! note
    Inn case a fatal application crash (panic) the crash log will end up on the stderr. Make sure to capture that as well for emergency debugging.
