<!--suppress HtmlDeprecatedAttribute -->
<h1 align="center">ContainerSSH UNIX Utilities Library</h1>

This library contains UNIX and Linux related utilities for ContainerSSH

<p align="center"><strong>⚠⚠⚠ Warning: This is a developer documentation. ⚠⚠⚠</strong><br />The user documentation for ContainerSSH is located at <a href="https://containerssh.io">containerssh.io</a>.</p>

## ParseCMD

The `ParseCMD()` method takes a command line and parses it into an execv-compatible slice of strings.

```go
args, err := unixutils.ParseCMD("/bin/sh -c 'echo \"Hello world!\"'")
//args will be: ["/bin/sh", "-c", "echo \"Hello world!\"]
```
