package ssh

type RequestType string

const (
	RequestTypeEnv       RequestType = "env"
	RequestTypePty       RequestType = "pty-req"
	RequestTypeShell     RequestType = "shell"
	RequestTypeExec      RequestType = "exec"
	RequestTypeSubsystem RequestType = "subsystem"
	RequestTypeWindow    RequestType = "window-change"
	RequestTypeSignal    RequestType = "signal"
)

const (
	SSH_MSG_USERAUTH_REQUEST = 50
)
