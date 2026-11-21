package ssh

type RequestType string

const (
	// Channel
	RequestTypeEnv          RequestType = "env"
	RequestTypePty          RequestType = "pty-req"
	RequestTypeShell        RequestType = "shell"
	RequestTypeExec         RequestType = "exec"
	RequestTypeSubsystem    RequestType = "subsystem"
	RequestTypeWindow       RequestType = "window-change"
	RequestTypeSignal       RequestType = "signal"
	RequestTypeX11          RequestType = "x11-req"

	// Global
	RequestTypeReverseForward           RequestType = "tcpip-forward"
	RequestTypeCancelReverseForward     RequestType = "cancel-tcpip-forward"
	RequestTypeStreamLocalForward       RequestType = "streamlocal-forward@openssh.com"
	RequestTypeCancelStreamLocalForward RequestType = "cancel-streamlocal-forward@openssh.com"
)

const (
	SSH_MSG_USERAUTH_REQUEST = 50
)
