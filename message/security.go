package message

// ESecurityExecRejected indicates that a program execution request has been rejected because it doesn't conform to the
// security settings.
const ESecurityExecRejected = "SECURITY_EXEC_REJECTED"

// ESecurityFailedSetEnv indicates that a program execution failed in conjunction with the forceCommand option because
// ContainerSSH could not set the `SSH_ORIGINAL_COMMAND` environment variable on the backend.
const ESecurityFailedSetEnv = "SECURITY_EXEC_FAILED_SETENV"

// MSecurityForcingCommand indicates that ContainerSSH is replacing the command passed from the client (if any) to the
// specified command and is setting the `SSH_ORIGINAL_COMMAND` environment variable.
const MSecurityForcingCommand = "SECURITY_EXEC_FORCING_COMMAND"

// ESecurityShellRejected indicates that ContainerSSH rejected launching a shell due to the security settings.
const ESecurityShellRejected = "SECURITY_SHELL_REJECTED"

// ESecuritySubsystemRejected indicates that ContainerSSH rejected the subsystem because it does pass the security
// settings.
const ESecuritySubsystemRejected = "SECURITY_SUBSYSTEM_REJECTED"

// ESecurityTTYRejected indicates that ContainerSSH rejected the pseudoterminal request because of the security
// settings.
const ESecurityTTYRejected = "SECURITY_TTY_REJECTED"

// ESecurityEnvRejected indicates that ContainerSSH rejected setting the environment variable because it does not pass
// the security settings.
const ESecurityEnvRejected = "SECURITY_ENV_REJECTED"

// ESecuritySignalRejected indicates that ContainerSSH rejected delivering a signal because it does not pass the
// security settings.
const ESecuritySignalRejected = "SECURITY_SIGNAL_REJECTED"

const ESecurityX11ForwardingRejected = "SECURITY_X11_FORWARDING_REJECTED"

// ESecurityMaxSessions indicates that the client has reached the maximum number of configured sessions, the new session
// request is therefore rejected.
const ESecurityMaxSessions = "SECURITY_MAX_SESSIONS"

const ESecurityForwardingRejected = "SECURITY_FORWARDING_REJECTED"

const ESecurityReverseForwardingRejected = "SECURITY_REVERSE_FORWARDING_REJECTED"