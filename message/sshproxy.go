package message

// The connection to the designated target server failed. If this message is logged on the debug level the connection will
// be retried. If this message is logged on the error level no more attempts will be made.
const ESSHProxyBackendConnectionFailed = "SSHPROXY_BACKEND_FAILED"

// The operation couldn't complete because the user already disconnected
const ESSHProxyDisconnected = "SSHPROXY_DISCONNECTED"

// The connection could not be established because the backend refused our authentication attempt. This is usually due
// to misconfigured credentials to the backend.
const ESSHProxyBackendHandshakeFailed = "SSHPROXY_BACKEND_HANDSHAKE_FAILED"

// ContainerSSH encountered an unexpected host key fingerprint on the backend while trying to proxy the connection.
// This is either due to a misconfiguration (not all host keys are listed), or a MITM attack between ContainerSSH and
// the target server.
const ESSHProxyInvalidFingerprint = "SSHPROXY_INVALID_FINGERPRINT"

// The client tried to perform an operation after the program has already been started which can only be performed
// before the program is started.
const ESSHProxyProgramAlreadyStarted = "SSHPROXY_PROGRAM_ALREADY_STARTED"

// The client tried to request an action that can only be performed once the program has started before the program
// started.
const ESSHProxyProgramNotStarted = "SSHPROXY_PROGRAM_NOT_STARTED"

// ContainerSSH failed to copy the stdin to the backing connection. This is usually due to an underlying network problem.
const ESSHProxyStdinError = "SSHPROXY_STDIN_ERROR"

// ContainerSSH failed to copy the stdout from the backing connection. This is usually due to an underlying network problem.
const ESSHProxyStdoutError = "SSHPROXY_STDOUT_ERROR"

// ContainerSSH failed to copy the stderr from the backing connection. This is usually due to an underlying network problem.
const ESSHProxyStderrError = "SSHPROXY_STDERR_ERROR"

// ContainerSSH is connecting the backing server.
const MSSHProxyConnecting = "SSHPROXY_CONNECTING"

// ContainerSSH failed to send a request to the backing server.
const ESSHProxyBackendRequestFailed = "SSHPROXY_BACKEND_REQUEST_FAILED"

// ContainerSSH failed to change the window size on the backend channel. This may be because of an underlying network
// issue, a policy-based rejection from the backend server, or a bug in the backend server.
const ESSHProxyWindowChangeFailed = "SSHPROXY_BACKEND_WINDOW_CHANGE_FAILED"

// ContainerSSH is shutting down and is sending TERM and KILL signals on the backend connection.
const MSSHProxyShutdown = "SSHPROXY_SHUTDOWN"

// ContainerSSH failed to deliver a signal on the backend channel. This may be because of an underlying network issue,
// a policy-based block on the backend server, or a general issue with the backend.
const ESSHProxySignalFailed = "SSHPROXY_BACKEND_SIGNAL_FAILED"

// The ContainerSSH SSH proxy module failed to close the client connection.
const ESSHProxyChannelCloseFailed = "SSHPROXY_CHANNEL_CLOSE_FAILED"

// ContainerSSH failed to open a session on the backend server with the SSH Proxy backend.
const ESSHProxyBackendSessionFailed = "SSHPROXY_BACKEND_SESSION_FAILED"

// The ContainerSSH SSH proxy backend is opening a new session.
const MSSHProxySession = "SSHPROXY_SESSION"

// The ContainerSSH SSH proxy backend has opened a new session.
const MSSHProxySessionOpen = "SSHPROXY_SESSION_OPEN"

// The ContainerSSH SSH proxy backend is closing a session.
const MSSHProxySessionClose = "SSHPROXY_SESSION_CLOSE"

// The ContainerSSH SSH proxy backend has closed a session.
const MSSHProxySessionClosed = "SSHPROXY_SESSION_CLOSED"

// ContainerSSH failed to close the session channel on the backend. This may be because of an underlying network issue
// or a problem with the backend server.
const ESSHProxySessionCloseFailed = "SSHPROXY_SESSION_CLOSE_FAILED"

// The ContainerSSH SSH proxy backend has received an exit-signal message.
const MSSHProxyExitSignal = "SSHPROXY_EXIT_SIGNAL"

// The ContainerSSH SSH proxy backend has received an exit-signal message, but failed to decode the message. This is
// most likely due to a bug in the backend SSH server.
const MSSHProxyExitSignalDecodeFailed = "SSHPROXY_EXIT_SIGNAL_DECODE_FAILED"

// The ContainerSSH SSH proxy backend has received an exit-status message.
const MSSHProxyExitStatus = "SSHPROXY_EXIT_STATUS"

// The ContainerSSH SSH proxy backend has received an exit-status message, but failed to decode the message. This is
// most likely due to a bug in the backend SSH server.
const MSSHProxyExitStatusDecodeFailed = "SSHPROXY_EXIT_STATUS_DECODE_FAILED"

// The ContainerSSH SSH proxy received a disconnect from the client.
const MSSHProxyDisconnected = "SSHPROXY_DISCONNECTED"

// The action cannot be performed because the connection is shutting down.
const ESSHProxyShuttingDown = "SSHPROXY_SHUTTING_DOWN"

// The backend closed the SSH session towards ContainerSSH.
const MSSHProxyBackendSessionClosed = "SSHPROXY_BACKEND_SESSION_CLOSED"

// ContainerSSH is closing the session towards the backend service.
const MSSHProxyBackendSessionClosing = "SSHPROXY_BACKEND_SESSION_CLOSING"

// ContainerSSH could not close the session towards the backend service.
const ESSHProxyBackendCloseFailed = "SSHPROXY_BACKEND_SESSION_CLOSE_FAILED"

// ContainerSSH is disconnecting from the backing server.
const MSSHProxyBackendDisconnecting = "SSHPROXY_BACKEND_DISCONNECTING"

const MSSHProxyBackendDisconnectFailed = "SSHPROXY_BACKEND_DISCONNECT_FAILED"

const MSSHProxyBackendDisconnected = "SSHPROXY_BACKEND_DISCONNECTED"

const MSSHProxyStderrComplete = "SSHPROXY_STDERR_COMPLETE"

const MSSHProxyStdoutComplete = "SSHPROXY_STDOUT_COMPLETE"
