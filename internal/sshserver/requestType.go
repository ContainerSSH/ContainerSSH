package sshserver

type requestType string

const (
	requestTypeEnv       requestType = "env"
	requestTypePty       requestType = "pty-req"
	requestTypeShell     requestType = "shell"
	requestTypeExec      requestType = "exec"
	requestTypeSubsystem requestType = "subsystem"
	requestTypeWindow    requestType = "window-change"
	requestTypeSignal    requestType = "signal"
)
