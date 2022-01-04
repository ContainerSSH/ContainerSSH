package ssh

import (
	"golang.org/x/crypto/ssh"
)

type EnvRequestPayload struct {
	Name  string
	Value string
}

type ExecRequestPayload struct {
	Exec string
}

type ExitSignalPayload struct {
	Signal       string
	CoreDumped   bool
	ErrorMessage string
	LanguageTag  string
}

type ExitStatusPayload struct {
	ExitStatus uint32
}

type PtyRequestPayload struct {
	Term     string
	Columns  uint32
	Rows     uint32
	Width    uint32
	Height   uint32
	ModeList []byte
}

type ShellRequestPayload struct {
}

type SignalRequestPayload struct {
	Signal string
}

type SubsystemRequestPayload struct {
	Subsystem string
}

type WindowRequestPayload struct {
	Columns uint32
	Rows    uint32
	Width   uint32
	Height  uint32
}

func Unmarshal(data []byte, out interface{}) error {
	return ssh.Unmarshal(data, out)
}
