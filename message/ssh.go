package message

// MSSHConnected indicates that a user has connected over SSH.
const MSSHConnected = "SSH_CONNECTED"

// MSSHDisconnected indicates that an SSH connection has been severed.
const MSSHDisconnected = "SSH_DISCONNECTED"

// ESSHHandshakeFailed indicates that the connecting party failed to establish a secure SSH connection. This is most
// likely due to invalid credentials or a backend error.
const ESSHHandshakeFailed = "SSH_HANDSHAKE_FAILED"

// MSSHHandshakeSuccessful indicates that the user has provided valid credentials and has now established an SSH
// connection.
const MSSHHandshakeSuccessful = "SSH_HANDSHAKE_SUCCESSFUL"

// ESSHUnsupportedGlobalRequest indicates that the user's client has sent a global request ContainerSSH does not support.
// This is nothing to worry about.
const ESSHUnsupportedGlobalRequest = "SSH_UNSUPPORTED_GLOBAL_REQUEST"

// ESSHKeepAliveFailed indicates that ContainerSSH couldn't send or didn't receive a response to a keepalive packet
const ESSHKeepAliveFailed = "SSH_KEEPALIVE_NORESP"

// ESSHReplyFailed indicates that ContainerSSH couldn't send the reply to a request to the user. This is usually the
// case if the user suddenly disconnects.
const ESSHReplyFailed = "SSH_REPLY_SEND_FAILED"

// ESSHUnsupportedChannelType indicates that the user requested a channel type that ContainerSSH doesn't support.
const ESSHUnsupportedChannelType = "SSH_UNSUPPORTED_CHANNEL_TYPE"

// ESSHAlreadyRunning indicates that the SSH server is already running and has been started again. This is a bug, please
// report it.
const ESSHAlreadyRunning = "SSH_ALREADY_RUNNING"

// ESSHStartFailed indicates that ContainerSSH failed to start the SSH service. This is usually because of invalid
// configuration.
const ESSHStartFailed = "SSH_START_FAILED"

// ESSHListenCloseFailed indicates that ContainerSSH failed to close the listen socket.
const ESSHListenCloseFailed = "SSH_LISTEN_CLOSE_FAILED"

// MSSHNewChannel indicates that a user has established a new SSH channel. (Not connection!)
const MSSHNewChannel = "SSH_NEW_CHANNEL"

// MSSHNewChannelRejected indicates that the user has requested a new channel to be opened, but was rejected.
const MSSHNewChannelRejected = "SSH_NEW_CHANNEL_REJECTED"

// MSSHServiceAvailable indicates that the SSH service is now online and ready for service.
const MSSHServiceAvailable = "SSH_AVAILABLE"

// ESSHAuthUnavailable indicates that the user has requested an authentication method that is currently unavailable.
const ESSHAuthUnavailable = "SSH_AUTH_UNAVAILABLE"

// ESSHAuthFailed indicates that the user has provided invalid credentials.
const ESSHAuthFailed = "SSH_AUTH_FAILED"

// ESSHAuthSuccessful indicates that the user has provided valid credentials and is now authenticated.
const ESSHAuthSuccessful = "SSH_AUTH_SUCCESSFUL"

// ESSHExitCodeFailed indicates that ContainerSSH failed to obtain and send the exit code of the program to the user.
const ESSHExitCodeFailed = "SSH_EXIT_CODE_FAILED"

// ESSHDecodeFailed indicates that ContainerSSH failed to decode something from the user. This is either a bug in
// ContainerSSH or in the connecting client.
const ESSHDecodeFailed = "SSH_DECODE_FAILED"

// MSSHExit indicates that ContainerSSH is sending the exit code of the program to the user.
const MSSHExit = "SSH_EXIT"

// MSSHExitSignal indicates that ContainerSSH is sending the exit signal from an abnormally exited program to the user.
const MSSHExitSignal = "SSH_EXIT_SIGNAL"

const MSSHGlobalRequest = "SSH_GLOBAL_REQUEST"

const MSSHGlobalRequestFailed = "SSH_GLOBAL_REQUEST_FAILED"

const MSSHGlobalRequestSuccessful = "SSH_GLOBAL_REQUEST_SUCCESSFUL"

// MSSHChannelRequest indicates that the user has sent a new channel-specific request.
const MSSHChannelRequest = "SSH_CHANNEL_REQUEST"

// MSSHChannelRequestFailed indicates that ContainerSSH couldn't fulfil the channel-specific request.
const MSSHChannelRequestFailed = "SSH_CHANNEL_REQUEST_FAILED"

// MSSHChannelRequestSuccessful indicates that ContainerSSH has successfully processed the channel-specific request.
const MSSHChannelRequestSuccessful = "SSH_CHANNEL_REQUEST_SUCCESSFUL"

// ESSHBackendRejected indicates that the backend has rejected the connecting user after successful authentication.
const ESSHBackendRejected = "SSH_BACKEND_REJECTED_HANDSHAKE"

// ESSHSOReuseFailed indicates that ContainerSSH failed to set the socket to reuse. This may cause ContainerSSH to fail
// on a restart.
const ESSHSOReuseFailed = "SSH_SOCKET_REUSE_FAILED"

// ESSHNotImplemented indicates that a feature is not implemented in the backend.
const ESSHNotImplemented = "SSH_NOT_IMPLEMENTED"
