package message

// ESSHProxyBackendConnectionFailed indicates that the connection to the designated target server failed. If this
// message is logged on the debug level the connection will be retried. If this message is logged on the error level no
// more attempts will be made.
const ESSHProxyBackendConnectionFailed = "SSHPROXY_BACKEND_FAILED"

// ESSHProxyDisconnected indicates that the operation couldn't complete because the user already disconnected.
const ESSHProxyDisconnected = "SSHPROXY_DISCONNECTED"

// ESSHProxyBackendHandshakeFailed indicates that the connection could not be established because the backend refused
// our authentication attempt. This is usually due to misconfigured credentials to the backend.
const ESSHProxyBackendHandshakeFailed = "SSHPROXY_BACKEND_HANDSHAKE_FAILED"

// ESSHProxyInvalidFingerprint indicates that ContainerSSH encountered an unexpected host key fingerprint on the backend
// while trying to proxy the connection. This is either due to a misconfiguration (not all host keys are listed), or a
// MITM attack between ContainerSSH and the target server.
const ESSHProxyInvalidFingerprint = "SSHPROXY_INVALID_FINGERPRINT"

// ESSHProxyProgramAlreadyStarted indicates that the client tried to perform an operation after the program has already
// been started which can only be performed before the program is started.
const ESSHProxyProgramAlreadyStarted = "SSHPROXY_PROGRAM_ALREADY_STARTED"

// ESSHProxyProgramNotStarted indicates that the client tried to request an action that can only be performed once the
// program has started before the program started.
const ESSHProxyProgramNotStarted = "SSHPROXY_PROGRAM_NOT_STARTED"

// ESSHProxyStdinError indicates that ContainerSSH failed to copy the stdin to the backing connection. This is usually
// due to an underlying network problem.
const ESSHProxyStdinError = "SSHPROXY_STDIN_ERROR"

// ESSHProxyStdoutError indicates that ContainerSSH failed to copy the stdout from the backing connection. This is
// usually due to an underlying network problem.
const ESSHProxyStdoutError = "SSHPROXY_STDOUT_ERROR"

// ESSHProxyStderrError indicates that ContainerSSH failed to copy the stderr from the backing connection. This is
// usually due to an underlying network problem.
const ESSHProxyStderrError = "SSHPROXY_STDERR_ERROR"

// MSSHProxyConnecting indicates that ContainerSSH is connecting the backing server.
const MSSHProxyConnecting = "SSHPROXY_CONNECTING"

// ESSHProxyBackendRequestFailed indicates that ContainerSSH failed to send a request to the backing server.
const ESSHProxyBackendRequestFailed = "SSHPROXY_BACKEND_REQUEST_FAILED"

// ESSHProxyWindowChangeFailed indicates that ContainerSSH failed to change the window size on the backend channel. This
// may be because of an underlying network issue, a policy-based rejection from the backend server, or a bug in the backend server.
const ESSHProxyWindowChangeFailed = "SSHPROXY_BACKEND_WINDOW_CHANGE_FAILED"

const ESSHProxyX11RequestFailed = "SSHPROXY_X11_FAILED"

// MSSHProxyShutdown indicates that ContainerSSH is shutting down and is sending TERM and KILL signals on the backend
// connection.
const MSSHProxyShutdown = "SSHPROXY_SHUTDOWN"

// ESSHProxySignalFailed indicates that ContainerSSH failed to deliver a signal on the backend channel. This may be
// because of an underlying network issue, a policy-based block on the backend server, or a general issue with the backend.
const ESSHProxySignalFailed = "SSHPROXY_BACKEND_SIGNAL_FAILED"

// ESSHProxyChannelCloseFailed indicates that the ContainerSSH SSH proxy module failed to close the client connection.
const ESSHProxyChannelCloseFailed = "SSHPROXY_CHANNEL_CLOSE_FAILED"

// ESSHProxyBackendSessionFailed indicates that ContainerSSH failed to open a session on the backend server with the SSH
// Proxy backend.
const ESSHProxyBackendSessionFailed = "SSHPROXY_BACKEND_SESSION_FAILED"

// ESSHProxyBackendForwardFailed indicates that ContainerSSH failed to open a forwarding channel on the backend server with the SSH Proxy backend
const ESSHProxyBackendForwardFailed = "SSHPROXY_BACKEND_FORWARD_FAILED"

// MSSHProxySession indicates that the ContainerSSH SSH proxy backend is opening a new session.
const MSSHProxySession = "SSHPROXY_SESSION"

// MSSHProxyForward indicates that the ContainerSSH proxy backend is opening a new proxy connection
const MSSHProxyForward = "SSHPROXY_TCP_FORWARD"

// MSSHProxySessionOpen indicates that the ContainerSSH SSH proxy backend has opened a new session.
const MSSHProxySessionOpen = "SSHPROXY_SESSION_OPEN"

// MSSHProxySessionClose indicates that the ContainerSSH SSH proxy backend is closing a session.
const MSSHProxySessionClose = "SSHPROXY_SESSION_CLOSE"

// MSSHProxySessionClosed indicates that the ContainerSSH SSH proxy backend has closed a session.
const MSSHProxySessionClosed = "SSHPROXY_SESSION_CLOSED"

// ESSHProxySessionCloseFailed indicates that ContainerSSH failed to close the session channel on the backend. This may
// be because of an underlying network issue or a problem with the backend server.
const ESSHProxySessionCloseFailed = "SSHPROXY_SESSION_CLOSE_FAILED"

// MSSHProxyExitSignal indicates that the ContainerSSH SSH proxy backend has received an exit-signal message.
const MSSHProxyExitSignal = "SSHPROXY_EXIT_SIGNAL"

// MSSHProxyExitSignalDecodeFailed indicates that the ContainerSSH SSH proxy backend has received an exit-signal
// message, but failed to decode the message. This is most likely due to a bug in the backend SSH server.
const MSSHProxyExitSignalDecodeFailed = "SSHPROXY_EXIT_SIGNAL_DECODE_FAILED"

// MSSHProxyExitStatus indicates that the ContainerSSH SSH proxy backend has received an exit-status message.
const MSSHProxyExitStatus = "SSHPROXY_EXIT_STATUS"

// MSSHProxyExitStatusDecodeFailed indicates that the ContainerSSH SSH proxy backend has received an exit-status
// message, but failed to decode the message. This is most likely due to a bug in the backend SSH server.
const MSSHProxyExitStatusDecodeFailed = "SSHPROXY_EXIT_STATUS_DECODE_FAILED"

// MSSHProxyDisconnected indicates that the ContainerSSH SSH proxy received a disconnect from the client.
const MSSHProxyDisconnected = "SSHPROXY_DISCONNECTED"

// ESSHProxyPayloadUnmarshalFailed indicates that the backend server has sent an invalid payload
const ESSHProxyPayloadUnmarshalFailed = "SSHPROXY_UNMARSHAL_FAILED"

// ESSHProxyShuttingDown indicates that the action cannot be performed because the connection is shutting down.
const ESSHProxyShuttingDown = "SSHPROXY_SHUTTING_DOWN"

// MSSHProxyBackendSessionClosed indicates that the backend closed the SSH session towards ContainerSSH.
const MSSHProxyBackendSessionClosed = "SSHPROXY_BACKEND_SESSION_CLOSED"

// MSSHProxyBackendSessionClosing indicates that ContainerSSH is closing the session towards the backend service.
const MSSHProxyBackendSessionClosing = "SSHPROXY_BACKEND_SESSION_CLOSING"

// ESSHProxyBackendCloseFailed indicates that ContainerSSH could not close the session towards the backend service.
const ESSHProxyBackendCloseFailed = "SSHPROXY_BACKEND_SESSION_CLOSE_FAILED"

// MSSHProxyBackendDisconnecting indicates that ContainerSSH is disconnecting from the backing server.
const MSSHProxyBackendDisconnecting = "SSHPROXY_BACKEND_DISCONNECTING"

// MSSHProxyBackendDisconnectFailed indicates that the SSH proxy backend failed to disconnect from the backing server.
const MSSHProxyBackendDisconnectFailed = "SSHPROXY_BACKEND_DISCONNECT_FAILED"

// MSSHProxyBackendDisconnected indicates that the SSH proxy backend disconnected from the backing server.
const MSSHProxyBackendDisconnected = "SSHPROXY_BACKEND_DISCONNECTED"

// MSSHProxyStderrComplete indicates that the SSH proxy backend finished streaming the standard error to the client.
const MSSHProxyStderrComplete = "SSHPROXY_STDERR_COMPLETE"

// MSSHProxyStdoutComplete indicates that the SSH proxy backend finished streaming the standard output to the client.
const MSSHProxyStdoutComplete = "SSHPROXY_STDOUT_COMPLETE"
