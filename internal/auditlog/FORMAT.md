# The ContainerSSH Binary Audit Log Format, version 1 (draft)

The ContainerSSH audit log is stored in [CBOR](https://cbor.io/) + GZIP format.

However, before GZIP decoding you must provide/decode the **file header**. The file header is 40 bytes long. The fist 32 bytes must contain the string `ContainerSSH-Auditlog` in UTF-8 encoding, the rest padded with 0 bytes. The last 8 bytes contain the audit log format version number as a 64 bit unsigned little endian integer.

```
Header {
    Magic    [32]byte  # Must always be ContainerSSH-Auditlog\0
    Version  uint64    # Little endian encoding
}
```

After the first 40 bytes you will have to GZIP-decode the rest of the file and then the CBOR format. The main element of the CBOR container is an *array of messages* where each message has the following format:

```
Message {
  # ConnectionID is an opaque ID of the connection. 
  ConnectionID  string
  # Timestamp is a nanosecond timestamp when the message was created. 
  Timestamp  int64
  # Type of the Payload object. 
  MessageType  int32
  # Payload is always a pointer to a payload object. 
  Payload  struct
  # ChannelID is a identifier for an SSH channel, if applicable. -1 otherwise. 
  ChannelID  *uint64
}
```

The audit log protocol has the following message types at this time:

| Message type ID | Name | Payload type |
|-----------------|------|--------------|
| 0 | Connect | [PayloadConnect](#PayloadConnect) |
| 1 | Disconnect | *none* |
| 100 | Password authentication | [PayloadAuthPassword](#PayloadAuthPassword) |
| 101 | Password authentication successful | [PayloadAuthPassword](#PayloadAuthPassword) |
| 102 | Password authentication failed | [PayloadAuthPassword](#PayloadAuthPassword) |
| 103 | Password authentication backend error | [PayloadAuthPasswordBackendError](#PayloadAuthPasswordBackendError) |
| 104 | Public key authentication | [PayloadAuthPubKey](#PayloadAuthPubKey) |
| 105 | Public key authentication successful | [PayloadAuthPubKey](#PayloadAuthPubKey) |
| 106 | Public key authentication failed | [PayloadAuthPubKey](#PayloadAuthPubKey) |
| 107 | Public key authentication backend error | [PayloadAuthPubKeyBackendError](#PayloadAuthPubKeyBackendError) |
| 108 | Keyboard-interactive authentication challenge | [PayloadAuthKeyboardInteractiveChallenge](#PayloadAuthKeyboardInteractiveChallenge) |
| 109 | Keyboard-interactive authentication answer | [PayloadAuthKeyboardInteractiveAnswer](#PayloadAuthKeyboardInteractiveAnswer) |
| 110 | Keyboard-interactive authentication failed | [PayloadAuthKeyboardInteractiveFailed](#PayloadAuthKeyboardInteractiveFailed) |
| 111 | Keyboard-interactive authentication backend error | [PayloadAuthKeyboardInteractiveBackendError](#PayloadAuthKeyboardInteractiveBackendError) |
| 200 | Unknown global request | [PayloadGlobalRequestUnknown](#PayloadGlobalRequestUnknown) |
| 300 | New channel request | [PayloadNewChannel](#PayloadNewChannel) |
| 301 | New channel successful | [PayloadNewChannelSuccessful](#PayloadNewChannelSuccessful) |
| 302 | New channel failed | [PayloadNewChannelFailed](#PayloadNewChannelFailed) |
| 400 | Unknown channel request | [PayloadChannelRequestUnknownType](#PayloadChannelRequestUnknownType) |
| 401 | Failed to decode channel request | [PayloadChannelRequestDecodeFailed](#PayloadChannelRequestDecodeFailed) |
| 402 | Set environment variable | [PayloadChannelRequestSetEnv](#PayloadChannelRequestSetEnv) |
| 403 | Execute program | [PayloadChannelRequestExec](#PayloadChannelRequestExec) |
| 404 | Request interactive terminal | [PayloadChannelRequestPty](#PayloadChannelRequestPty) |
| 405 | Run shell | [PayloadChannelRequestShell](#PayloadChannelRequestShell) |
| 406 | Send signal to running process | [PayloadChannelRequestSignal](#PayloadChannelRequestSignal) |
| 407 | Request subsystem | [PayloadChannelRequestSubsystem](#PayloadChannelRequestSubsystem) |
| 408 | Change window size | [PayloadChannelRequestWindow](#PayloadChannelRequestWindow) |
| 496 | Close channel for writing | *none* |
| 497 | Close channel | *none* |
| 498 | Program exited with signal | [PayloadExitSignal](#PayloadExitSignal) |
| 499 | Program exited | [PayloadExit](#PayloadExit) |
| 500 | I/O | [PayloadIO](#PayloadIO) |
| 501 | Request failed | [PayloadRequestFailed](#PayloadRequestFailed) |

## PayloadConnect

PayloadConnect is the payload for TypeConnect messages. 

```
PayloadConnect {
  RemoteAddr  string  # RemoteAddr contains the IP address of the connecting user. 
  Country     string  # Country contains the country code looked up from the IP address. Contains "XX" if the lookup failed. 
}
```

## PayloadAuthPassword

PayloadAuthPassword is a payload for a message that indicates an authentication attempt, successful, or failed authentication. 

```
PayloadAuthPassword {
  Username  string
  Password  []uint8
}
```

## PayloadAuthPassword

PayloadAuthPassword is a payload for a message that indicates an authentication attempt, successful, or failed authentication. 

```
PayloadAuthPassword {
  Username  string
  Password  []uint8
}
```

## PayloadAuthPassword

PayloadAuthPassword is a payload for a message that indicates an authentication attempt, successful, or failed authentication. 

```
PayloadAuthPassword {
  Username  string
  Password  []uint8
}
```

## PayloadAuthPasswordBackendError

PayloadAuthPasswordBackendError is a payload for a message that indicates a backend failure during authentication. 

```
PayloadAuthPasswordBackendError {
  Username  string
  Password  []uint8
  Reason    string
}
```

## PayloadAuthPubKey

PayloadAuthPubKey is a payload for a public key based authentication 

```
PayloadAuthPubKey {
  Username  string
  Key       string
}
```

## PayloadAuthPubKey

PayloadAuthPubKey is a payload for a public key based authentication 

```
PayloadAuthPubKey {
  Username  string
  Key       string
}
```

## PayloadAuthPubKey

PayloadAuthPubKey is a payload for a public key based authentication 

```
PayloadAuthPubKey {
  Username  string
  Key       string
}
```

## PayloadAuthPubKeyBackendError

PayloadAuthPubKeyBackendError is a payload for a message indicating that there was a backend error while authenticating with public key. 

```
PayloadAuthPubKeyBackendError {
  Username  string
  Key       string
  Reason    string
}
```

## PayloadAuthKeyboardInteractiveChallenge

PayloadAuthKeyboardInteractiveChallenge is a message that indicates that a keyboard-interactive challenge has been sent to the user. Multiple challenge-response interactions can take place. 

```
PayloadAuthKeyboardInteractiveChallenge {
  Username     string
  Instruction  string
  Questions    []struct
}
```

## PayloadAuthKeyboardInteractiveAnswer

PayloadAuthKeyboardInteractiveAnswer is a message that indicates a response to a keyboard-interactive challenge. 

```
PayloadAuthKeyboardInteractiveAnswer {
  Username  string
  Answers   []struct
}
```

## PayloadAuthKeyboardInteractiveFailed

PayloadAuthKeyboardInteractiveFailed indicates that a keyboard-interactive authentication process has failed. 

```
PayloadAuthKeyboardInteractiveFailed {
  Username  string
}
```

## PayloadAuthKeyboardInteractiveBackendError

PayloadAuthKeyboardInteractiveBackendError indicates an error in the authentication backend during a keyboard-interactive authentication. 

```
PayloadAuthKeyboardInteractiveBackendError {
  Username  string
  Reason    string
}
```

## PayloadGlobalRequestUnknown

PayloadGlobalRequestUnknown Is a payload for the TypeGlobalRequestUnknown messages. 

```
PayloadGlobalRequestUnknown {
  RequestType  string
}
```

## PayloadNewChannel

PayloadNewChannel is a payload that signals a request for a new SSH channel 

```
PayloadNewChannel {
  ChannelType  string
}
```

## PayloadNewChannelSuccessful

PayloadNewChannelSuccessful is a payload that signals that a channel request was successful. 

```
PayloadNewChannelSuccessful {
  ChannelType  string
}
```

## PayloadNewChannelFailed

PayloadNewChannelFailed is a payload that signals that a request for a new channel has failed. 

```
PayloadNewChannelFailed {
  ChannelType  string
  Reason       string
}
```

## PayloadChannelRequestUnknownType

PayloadChannelRequestUnknownType is a payload signaling that a channel request was not supported. 

```
PayloadChannelRequestUnknownType {
  RequestID    uint64
  RequestType  string
  Payload      []uint8
}
```

## PayloadChannelRequestDecodeFailed

PayloadChannelRequestDecodeFailed is a payload that signals a supported request that the server was unable to decode. 

```
PayloadChannelRequestDecodeFailed {
  RequestID    uint64
  RequestType  string
  Payload      []uint8
  Reason       string
}
```

## PayloadChannelRequestSetEnv

PayloadChannelRequestSetEnv is a payload signaling the request for an environment variable. 

```
PayloadChannelRequestSetEnv {
  RequestID  uint64
  Name       string
  Value      string
}
```

## PayloadChannelRequestExec

PayloadChannelRequestExec is a payload signaling the request to execute a program. 

```
PayloadChannelRequestExec {
  RequestID  uint64
  Program    string
}
```

## PayloadChannelRequestPty

PayloadChannelRequestPty is a payload signaling the request for an interactive terminal. 

```
PayloadChannelRequestPty {
  RequestID  uint64
  Term       string
  Columns    uint32
  Rows       uint32
  Width      uint32
  Height     uint32
  ModeList   []uint8
}
```

## PayloadChannelRequestShell

PayloadChannelRequestShell is a payload signaling a request for a shell. 

```
PayloadChannelRequestShell {
  RequestID  uint64
}
```

## PayloadChannelRequestSignal

PayloadChannelRequestSignal is a payload signaling a signal request to be sent to the currently running program. 

```
PayloadChannelRequestSignal {
  RequestID  uint64
  Signal     string
}
```

## PayloadChannelRequestSubsystem

PayloadChannelRequestSubsystem is a payload requesting a well-known subsystem (e.g. sftp) 

```
PayloadChannelRequestSubsystem {
  RequestID  uint64
  Subsystem  string
}
```

## PayloadChannelRequestWindow

PayloadChannelRequestWindow is a payload requesting the change in the terminal window size. 

```
PayloadChannelRequestWindow {
  RequestID  uint64
  Columns    uint32
  Rows       uint32
  Width      uint32
  Height     uint32
}
```

## PayloadExitSignal

PayloadExitSignal indicates the signal that caused a program to abort. 

```
PayloadExitSignal {
  Signal        string
  CoreDumped    bool
  ErrorMessage  string
  LanguageTag   string
}
```

## PayloadExit

PayloadExit is the payload for a message that is sent when a program exits. 

```
PayloadExit {
  ExitStatus  uint32
}
```

## PayloadIO

PayloadIO The payload for I/O message types containing the data stream from/to the application. 

```
PayloadIO {
  Stream  uint  # 0 = stdin, 1 = stdout, 2 = stderr 
  Data    []uint8
}
```

## PayloadRequestFailed

PayloadRequestFailed is the payload for the TypeRequestFailed messages. 

```
PayloadRequestFailed {
  RequestID  uint64
  Reason     string
}
```

