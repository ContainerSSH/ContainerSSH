# The ContainerSSH Binary Audit Log Format, version 2 (draft)

The ContainerSSH audit log is stored in [CBOR](https://cbor.io/) + GZIP format.

However, before GZIP decoding you must provide/decode the **file header**. The file header is 40 bytes long. The fist 32 bytes must contain the string `ContainerSSH-Auditlog` in UTF-8 encoding, the rest padded with 0 bytes. The last 8 bytes contain the audit log format version number as a 64 bit unsigned little endian integer.

```
Header {
    Magic    [32]byte  # Must always be ContainerSSH-Auditlog\0
    Version  uint64    # Little endian encoding
}
```

After the first 40 bytes you will have to GZIP-decode the rest of the file. The result is a sequence of CBOR-encoded messages, one after the other, with no enclosing array (version 1 wrapped all messages in a single CBOR array). Decode messages until the stream ends. The GZIP stream may lack its end marker, so some GZIP readers report an unexpected end of file, and a log that is still being written may end in the middle of a message. Treat both as the end of the log. Each message is a CBOR map with the following keys:

```
Message {
  # ConnectionID is an opaque ID of the connection. 
  connectionId  string
  # Timestamp is a nanosecond timestamp when the message was created. 
  timestamp  int64
  # Type of the Payload object. 
  type  int32
  # Payload is always a pointer to a payload object. 
  payload  struct
  # ChannelID is an identifier for an SSH channel, if applicable. Null otherwise. 
  channelId  *uint64
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
| 198 | Handshake failed | [PayloadHandshakeFailed](#PayloadHandshakeFailed) |
| 199 | Handshake successful | [PayloadHandshakeSuccessful](#PayloadHandshakeSuccessful) |
| 200 | Unknown global request | [PayloadGlobalRequestUnknown](#PayloadGlobalRequestUnknown) |
| 201 | Failed to decode global request | [PayloadGlobalRequestDecodeFailed](#PayloadGlobalRequestDecodeFailed) |
| 202 | Request reverse port forwarding | [PayloadRequestReverseForward](#PayloadRequestReverseForward) |
| 203 | Request to stop reverse port forwarding | [PayloadRequestReverseForward](#PayloadRequestReverseForward) |
| 204 | Request reverse socket forwarding | [PayloadRequestStreamLocal](#PayloadRequestStreamLocal) |
| 205 | Request to stop reverse socket forwarding | [PayloadRequestStreamLocal](#PayloadRequestStreamLocal) |
| 300 | New channel request | [PayloadNewChannel](#PayloadNewChannel) |
| 301 | New channel successful | [PayloadNewChannelSuccessful](#PayloadNewChannelSuccessful) |
| 302 | New channel failed | [PayloadNewChannelFailed](#PayloadNewChannelFailed) |
| 303 | New client-to-server port forwarding channel | [PayloadNewForwardChannel](#PayloadNewForwardChannel) |
| 304 | New server-to-client port forwarding channel | [PayloadNewReverseForwardChannel](#PayloadNewReverseForwardChannel) |
| 305 | New server-to-client X11 forwarding channel | [PayloadNewReverseX11ForwardChannel](#PayloadNewReverseX11ForwardChannel) |
| 306 | New client-to-server unix socket forwarding channel | [PayloadRequestStreamLocal](#PayloadRequestStreamLocal) |
| 307 | New server-to-client unix socket forwarding channel | [PayloadRequestStreamLocal](#PayloadRequestStreamLocal) |
| 400 | Unknown channel request | [PayloadChannelRequestUnknownType](#PayloadChannelRequestUnknownType) |
| 401 | Failed to decode channel request | [PayloadChannelRequestDecodeFailed](#PayloadChannelRequestDecodeFailed) |
| 402 | Set environment variable | [PayloadChannelRequestSetEnv](#PayloadChannelRequestSetEnv) |
| 403 | Execute program | [PayloadChannelRequestExec](#PayloadChannelRequestExec) |
| 404 | Request interactive terminal | [PayloadChannelRequestPty](#PayloadChannelRequestPty) |
| 405 | Run shell | [PayloadChannelRequestShell](#PayloadChannelRequestShell) |
| 406 | Send signal to running process | [PayloadChannelRequestSignal](#PayloadChannelRequestSignal) |
| 407 | Request subsystem | [PayloadChannelRequestSubsystem](#PayloadChannelRequestSubsystem) |
| 408 | Change window size | [PayloadChannelRequestWindow](#PayloadChannelRequestWindow) |
| 409 | Request X11 forwarding | [PayloadChannelRequestX11](#PayloadChannelRequestX11) |
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
  remoteAddr  string  # RemoteAddr contains the IP address of the connecting user. 
  country     string  # Country contains the country code looked up from the IP address. Contains "XX" if the lookup failed. 
}
```

## PayloadAuthPassword

PayloadAuthPassword is a payload for a message that indicates an authentication attempt, successful, or failed authentication. 

```
PayloadAuthPassword {
  username  string
  password  []uint8
}
```

## PayloadAuthPassword

PayloadAuthPassword is a payload for a message that indicates an authentication attempt, successful, or failed authentication. 

```
PayloadAuthPassword {
  username  string
  password  []uint8
}
```

## PayloadAuthPassword

PayloadAuthPassword is a payload for a message that indicates an authentication attempt, successful, or failed authentication. 

```
PayloadAuthPassword {
  username  string
  password  []uint8
}
```

## PayloadAuthPasswordBackendError

PayloadAuthPasswordBackendError is a payload for a message that indicates a backend failure during authentication. 

```
PayloadAuthPasswordBackendError {
  username  string
  password  []uint8
  reason    string
}
```

## PayloadAuthPubKey

PayloadAuthPubKey is a payload for a public key based authentication. 

```
PayloadAuthPubKey {
  username  string
  key       string
}
```

## PayloadAuthPubKey

PayloadAuthPubKey is a payload for a public key based authentication. 

```
PayloadAuthPubKey {
  username  string
  key       string
}
```

## PayloadAuthPubKey

PayloadAuthPubKey is a payload for a public key based authentication. 

```
PayloadAuthPubKey {
  username  string
  key       string
}
```

## PayloadAuthPubKeyBackendError

PayloadAuthPubKeyBackendError is a payload for a message indicating that there was a backend error while authenticating with public key. 

```
PayloadAuthPubKeyBackendError {
  username  string
  key       string
  reason    string
}
```

## PayloadAuthKeyboardInteractiveChallenge

PayloadAuthKeyboardInteractiveChallenge is a message that indicates that a keyboard-interactive challenge has been sent to the user. Multiple challenge-response interactions can take place. 

```
PayloadAuthKeyboardInteractiveChallenge {
  username     string
  instruction  string
  questions    []struct
}
```

## PayloadAuthKeyboardInteractiveAnswer

PayloadAuthKeyboardInteractiveAnswer is a message that indicates a response to a keyboard-interactive challenge. 

```
PayloadAuthKeyboardInteractiveAnswer {
  username  string
  answers   []struct
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
  username  string
  reason    string
}
```

## PayloadHandshakeFailed

PayloadHandshakeFailed is a payload for a failed handshake. 

```
PayloadHandshakeFailed {
  reason  string
}
```

## PayloadHandshakeSuccessful

PayloadHandshakeSuccessful is a payload for a successful handshake. 

```
PayloadHandshakeSuccessful {
  username  string
}
```

## PayloadGlobalRequestUnknown

PayloadGlobalRequestUnknown Is a payload for the TypeGlobalRequestUnknown messages. 

```
PayloadGlobalRequestUnknown {
  requestType  string
}
```

## PayloadGlobalRequestDecodeFailed

PayloadGlobalRequestDecodeFailed is a payload that signals a supported request that the server was unable to decode 

```
PayloadGlobalRequestDecodeFailed {
  requestId    uint64
  requestType  string
  payload      []uint8
  reason       string
}
```

## PayloadRequestReverseForward

```
PayloadRequestReverseForward {
  BindHost  string
  BindPort  uint32
}
```

## PayloadRequestReverseForward

```
PayloadRequestReverseForward {
  BindHost  string
  BindPort  uint32
}
```

## PayloadRequestStreamLocal

```
PayloadRequestStreamLocal {
  Path  string
}
```

## PayloadRequestStreamLocal

```
PayloadRequestStreamLocal {
  Path  string
}
```

## PayloadNewChannel

PayloadNewChannel is a payload that signals a request for a new SSH channel 

```
PayloadNewChannel {
  channelType  string
}
```

## PayloadNewChannelSuccessful

PayloadNewChannelSuccessful is a payload that signals that a channel request was successful. 

```
PayloadNewChannelSuccessful {
  channelType  string
}
```

## PayloadNewChannelFailed

PayloadNewChannelFailed is a payload that signals that a request for a new channel has failed. 

```
PayloadNewChannelFailed {
  channelType  string
  reason       string
}
```

## PayloadNewForwardChannel

```
PayloadNewForwardChannel {
  HostToConnect   string
  PortToConnect   uint32
  OriginatorHost  string
  OriginatorPort  uint32
}
```

## PayloadNewReverseForwardChannel

```
PayloadNewReverseForwardChannel {
  ConnectedHost   string
  ConnectedPort   uint32
  OriginatorHost  string
  OriginatorPort  uint32
}
```

## PayloadNewReverseX11ForwardChannel

```
PayloadNewReverseX11ForwardChannel {
  OriginatorHost  string
  OriginatorPort  uint32
}
```

## PayloadRequestStreamLocal

```
PayloadRequestStreamLocal {
  Path  string
}
```

## PayloadRequestStreamLocal

```
PayloadRequestStreamLocal {
  Path  string
}
```

## PayloadChannelRequestUnknownType

PayloadChannelRequestUnknownType is a payload signaling that a channel request was not supported. 

```
PayloadChannelRequestUnknownType {
  requestId    uint64
  requestType  string
  payload      []uint8
}
```

## PayloadChannelRequestDecodeFailed

PayloadChannelRequestDecodeFailed is a payload that signals a supported request that the server was unable to decode. 

```
PayloadChannelRequestDecodeFailed {
  requestId    uint64
  requestType  string
  payload      []uint8
  reason       string
}
```

## PayloadChannelRequestSetEnv

PayloadChannelRequestSetEnv is a payload signaling the request for an environment variable. 

```
PayloadChannelRequestSetEnv {
  requestId  uint64
  name       string
  value      string
}
```

## PayloadChannelRequestExec

PayloadChannelRequestExec is a payload signaling the request to execute a program. 

```
PayloadChannelRequestExec {
  requestId  uint64
  program    string
}
```

## PayloadChannelRequestPty

PayloadChannelRequestPty is a payload signaling the request for an interactive terminal. 

```
PayloadChannelRequestPty {
  requestId  uint64
  term       string
  columns    uint32
  rows       uint32
  width      uint32
  height     uint32
  modelist   []uint8
}
```

## PayloadChannelRequestShell

PayloadChannelRequestShell is a payload signaling a request for a shell. 

```
PayloadChannelRequestShell {
  requestId  uint64
}
```

## PayloadChannelRequestSignal

PayloadChannelRequestSignal is a payload signaling a signal request to be sent to the currently running program. 

```
PayloadChannelRequestSignal {
  requestId  uint64
  signal     string
}
```

## PayloadChannelRequestSubsystem

PayloadChannelRequestSubsystem is a payload requesting a well-known subsystem (e.g. sftp) 

```
PayloadChannelRequestSubsystem {
  requestId  uint64
  subsystem  string
}
```

## PayloadChannelRequestWindow

PayloadChannelRequestWindow is a payload requesting the change in the terminal window size. 

```
PayloadChannelRequestWindow {
  requestId  uint64
  columns    uint32
  rows       uint32
  width      uint32
  height     uint32
}
```

## PayloadChannelRequestX11

```
PayloadChannelRequestX11 {
  requestId         uint64
  singleConnection  bool
  authProtocol      string
  cookie            string
  screen            uint32
}
```

## PayloadExitSignal

PayloadExitSignal indicates the signal that caused a program to abort. 

```
PayloadExitSignal {
  signal        string
  coreDumped    bool
  errorMessage  string
  languageTag   string
}
```

## PayloadExit

PayloadExit is the payload for a message that is sent when a program exits. 

```
PayloadExit {
  exitStatus  uint32
}
```

## PayloadIO

PayloadIO The payload for I/O message types containing the testdata stream from/to the application. 

```
PayloadIO {
  stream  uint  # 0 = stdin, 1 = stdout, 2 = stderr 
  data    []uint8
}
```

## PayloadRequestFailed

PayloadRequestFailed is the payload for the TypeRequestFailed messages. 

```
PayloadRequestFailed {
  requestId  uint64
  reason     string
}
```

