package message

// A user has connected over SSH.
const MSSHConnected = "SSH_CONNECTED"

// An SSH connection has been severed.
const MSSHDisconnected = "SSH_DISCONNECTED"

// The connecting party failed to establish a secure SSH connection. This is most likely due to invalid credentials
// or a backend error.
const ESSHHandshakeFailed = "SSH_HANDSHAKE_FAILED"

// The user has provided valid credentials and has now established an SSH connection.
const MSSHHandshakeSuccessful = "SSH_HANDSHAKE_SUCCESSFUL"

// The users client has send a global request ContainerSSH does not support. This is nothing to worry about.
const ESSHUnsupportedGlobalRequest = "SSH_UNSUPPORTED_GLOBAL_REQUEST"

// ContainerSSH couldn't send the reply to a request to the user. This is usually the case if the user suddenly
// disconnects.
const ESSHReplyFailed = "SSH_REPLY_SEND_FAILED"

// The user requested a channel type that ContainerSSH doesn't support (e.g. TCP/IP forwarding).
const ESSHUnsupportedChannelType = "SSH_UNSUPPORTED_CHANNEL_TYPE"

// The SSH server is already running and has been started again. This is a bug, please report it.
const ESSHAlreadyRunning = "SSH_ALREADY_RUNNING"

// ContainerSSH failed to start the SSH service. This is usually because of invalid configuration.
const ESSHStartFailed = "SSH_START_FAILED"

// ContainerSSH failed to close the listen socket.
const ESSHListenCloseFailed = "SSH_LISTEN_CLOSE_FAILED"

// A user has established a new SSH channel. (Not connection!)
const MSSHNewChannel = "SSH_NEW_CHANNEL"

// The user has requested a new channel to be opened, but was rejected.
const MSSHNewChannelRejected = "SSH_NEW_CHANNEL_REJECTED"

// The SSH service is now online and ready for service.
const MSSHServiceAvailable = "SSH_AVAILABLE"

// The user has requested an authentication method that is currently unavailable.
const ESSHAuthUnavailable = "SSH_AUTH_UNAVAILABLE"

// The user has provided invalid credentials.
const ESSHAuthFailed = "SSH_AUTH_FAILED"

// The user has provided valid credentials and is now authenticated.
const ESSHAuthSuccessful = "SSH_AUTH_SUCCESSFUL"

// ContainerSSH failed to obtain and send the exit code of the program to the user.
const ESSHExitCodeFailed = "SSH_EXIT_CODE_FAILED"

// ContainerSSH failed to decode something from the user. This is either a bug in ContainerSSH or in the connecting
// client.
const ESSHDecodeFailed = "SSH_DECODE_FAILED"

// ContainerSSH is sending the exit code of the program to the user.
const MSSHExit = "SSH_EXIT"

// ContainerSSH is sending the exit signal from an abnormally exited program to the user.
const MSSHExitSignal = "SSH_EXIT_SIGNAL"

// The user has send a new channel-specific request.
const MSSHChannelRequest = "SSH_CHANNEL_REQUEST"

// ContainerSSH couldn't fulfil the channel-specific request.
const MSSHChannelRequestFailed = "SSH_CHANNEL_REQUEST_FAILED"

// ContainerSSH has successfully processed the channel-specific request.
const MSSHChannelRequestSuccessful = "SSH_CHANNEL_REQUEST_SUCCESSFUL"

// The backend has rejected the connecting user after successful authentication.
const ESSHBackendRejected = "SSH_BACKEND_REJECTED_HANDSHAKE"

// ContainerSSH failed to set the socket to reuse. This may cause ContainerSSH to fail on a restart.
const ESSHSOReuseFailed = "SSH_SOCKET_REUSE_FAILED"

// This message indicates that a feature is not implemented in the backend.
const ESSHNotImplemented = "SSH_NOT_IMPLEMENTED"
