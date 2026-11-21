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

type ForwardTCPIPRequestPayload struct {
	Address string
	Port    uint32
}

type ForwardTCPChannelOpenPayload struct {
	ConnectedAddress  string
	ConnectedPort     uint32
	OriginatorAddress string
	OriginatorPort    uint32
}

type X11RequestPayload struct {
	SingleConnection bool
	Protocol         string
	Cookie           string
	Screen           uint32
}

type X11ChanOpenRequestPayload struct {
	OriginatorAddress string
	OriginatorPort    uint32
}

type DirectStreamLocalChannelOpenPayload struct {
	SocketPath string
	Reserved1  string
	Reserved2  uint32
}

type ForwardedStreamLocalChannelOpenPayload struct {
	SocketPath string
	Reserved   string
}

type StreamLocalForwardRequestPayload struct {
	//State uint32 this field is there in the docs but not actually present in the channel data, weird
	SocketPath string
}

func Unmarshal(data []byte, out interface{}) error {
	return ssh.Unmarshal(data, out)
}
