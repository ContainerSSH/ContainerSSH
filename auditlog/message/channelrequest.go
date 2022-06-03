package message

import (
	"bytes"
)

// PayloadChannelRequestUnknownType is a payload signaling that a channel request was not supported.
type PayloadChannelRequestUnknownType struct {
	RequestID uint64 `json:"requestId" yaml:"requestId"`

	RequestType string `json:"requestType" yaml:"requestType"`
	Payload     []byte `json:"payload" yaml:"payload"`
}

// Equals compares two PayloadChannelRequestUnknownType payloads.
func (p PayloadChannelRequestUnknownType) Equals(other Payload) bool {
	p2, ok := other.(PayloadChannelRequestUnknownType)
	if !ok {
		return false
	}
	return p.RequestID == p2.RequestID && p.RequestType == p2.RequestType
}

// PayloadChannelRequestDecodeFailed is a payload that signals a supported request that the server was unable to decode.
type PayloadChannelRequestDecodeFailed struct {
	RequestID uint64 `json:"requestId" yaml:"requestId"`

	RequestType string `json:"requestType" yaml:"requestType"`
	Payload     []byte `json:"payload" yaml:"payload"`
	Reason      string `json:"reason" yaml:"reason"`
}

// Equals compares two PayloadChannelRequestDecodeFailed payloads.
func (p PayloadChannelRequestDecodeFailed) Equals(other Payload) bool {
	p2, ok := other.(PayloadChannelRequestDecodeFailed)
	if !ok {
		return false
	}
	return p.RequestID == p2.RequestID && p.RequestType == p2.RequestType && p.Reason == p2.Reason
}

// PayloadChannelRequestSetEnv is a payload signaling the request for an environment variable.
type PayloadChannelRequestSetEnv struct {
	RequestID uint64 `json:"requestId" yaml:"requestId"`

	Name  string `json:"name" yaml:"name"`
	Value string `json:"value" yaml:"value"`
}

// Equals compares two PayloadChannelRequestSetEnv payloads.
func (p PayloadChannelRequestSetEnv) Equals(other Payload) bool {
	p2, ok := other.(PayloadChannelRequestSetEnv)
	if !ok {
		return false
	}
	return p.RequestID == p2.RequestID && p.Name == p2.Name && p.Value == p2.Value
}

// PayloadChannelRequestExec is a payload signaling the request to execute a program.
type PayloadChannelRequestExec struct {
	RequestID uint64 `json:"requestId" yaml:"requestId"`

	Program string `json:"program" yaml:"program"`
}

// Equals compares two PayloadChannelRequestExec payloads.
func (p PayloadChannelRequestExec) Equals(other Payload) bool {
	p2, ok := other.(PayloadChannelRequestExec)
	if !ok {
		return false
	}
	return p.RequestID == p2.RequestID && p.Program == p2.Program
}

// PayloadChannelRequestPty is a payload signaling the request for an interactive terminal.
type PayloadChannelRequestPty struct {
	RequestID uint64 `json:"requestId" yaml:"requestId"`

	Term     string `json:"term" yaml:"term"`
	Columns  uint32 `json:"columns" yaml:"columns"`
	Rows     uint32 `json:"rows" yaml:"rows"`
	Width    uint32 `json:"width" yaml:"width"`
	Height   uint32 `json:"height" yaml:"height"`
	ModeList []byte `json:"modelist" yaml:"modelist"`
}

// Equals compares two PayloadChannelRequestPty payloads.
func (p PayloadChannelRequestPty) Equals(other Payload) bool {
	p2, ok := other.(PayloadChannelRequestPty)
	if !ok {
		return false
	}
	return p.RequestID == p2.RequestID &&
		p.Term == p2.Term &&
		p.Columns == p2.Columns &&
		p.Rows == p2.Rows &&
		p.Width == p2.Width &&
		p.Height == p2.Height &&
		bytes.Equal(p.ModeList, p2.ModeList)
}

type PayloadChannelRequestX11 struct {
	RequestID        uint64 `json:"requestId" yaml:"requestId"`

	SingleConnection bool   `json:"singleConnection" yaml:"singleConnection"`
	AuthProtocol     string `json:"authProtocol" yaml:"authProtocol"`
	Cookie           string `json:"cookie" yaml:"cookie"`
	Screen           uint32 `json:"screen" yaml:"screen"`
}

func (p PayloadChannelRequestX11) Equals(other Payload) bool {
	p2, ok := other.(PayloadChannelRequestX11)
	if !ok {
		return false
	}
	return p.RequestID == p2.RequestID &&
		p.SingleConnection == p2.SingleConnection &&
		p.AuthProtocol == p2.AuthProtocol &&
		p.Cookie == p2.Cookie &&
		p.Screen == p2.Screen
}

// PayloadChannelRequestShell is a payload signaling a request for a shell.
type PayloadChannelRequestShell struct {
	RequestID uint64 `json:"requestId" yaml:"requestId"`
}

// Equals compares two PayloadChannelRequestShell payloads.
func (p PayloadChannelRequestShell) Equals(other Payload) bool {
	p2, ok := other.(PayloadChannelRequestShell)
	return ok && p.RequestID == p2.RequestID
}

// PayloadChannelRequestSignal is a payload signaling a signal request to be sent to the currently running program.
type PayloadChannelRequestSignal struct {
	RequestID uint64 `json:"requestId" yaml:"requestId"`

	Signal string `json:"signal" yaml:"signal"`
}

// Equals compares two PayloadChannelRequestSignal payloads.
func (p PayloadChannelRequestSignal) Equals(other Payload) bool {
	p2, ok := other.(PayloadChannelRequestSignal)
	if !ok {
		return false
	}
	return p.RequestID == p2.RequestID && p.Signal == p2.Signal
}

// PayloadChannelRequestSubsystem is a payload requesting a well-known subsystem (e.g. sftp)
type PayloadChannelRequestSubsystem struct {
	RequestID uint64 `json:"requestId" yaml:"requestId"`

	Subsystem string `json:"subsystem" yaml:"subsystem"`
}

// Equals compares two PayloadChannelRequestSubsystem payloads.
func (p PayloadChannelRequestSubsystem) Equals(other Payload) bool {
	p2, ok := other.(PayloadChannelRequestSubsystem)
	if !ok {
		return false
	}
	return p.RequestID == p2.RequestID && p.Subsystem == p2.Subsystem
}

// PayloadChannelRequestWindow is a payload requesting the change in the terminal window size.
type PayloadChannelRequestWindow struct {
	RequestID uint64 `json:"requestId" yaml:"requestId"`
	Columns   uint32 `json:"columns" yaml:"columns"`
	Rows      uint32 `json:"rows" yaml:"rows"`
	Width     uint32 `json:"width" yaml:"width"`
	Height    uint32 `json:"height" yaml:"height"`
}

// Equals compares two PayloadChannelRequestWindow payloads.
func (p PayloadChannelRequestWindow) Equals(other Payload) bool {
	p2, ok := other.(PayloadChannelRequestWindow)
	if !ok {
		return false
	}
	return p.RequestID == p2.RequestID && p.Columns == p2.Columns && p.Rows == p2.Rows && p.Width == p2.Width &&
		p.Height == p2.Height
}

// PayloadExit is the payload for a message that is sent when a program exits.
type PayloadExit struct {
	ExitStatus uint32 `json:"exitStatus" yaml:"exitStatus"`
}

// Equals compares two PayloadExit payloads.
func (p PayloadExit) Equals(other Payload) bool {
	p2, ok := other.(PayloadExit)
	if !ok {
		return false
	}
	return p.ExitStatus == p2.ExitStatus
}

// PayloadExitSignal indicates the signal that caused a program to abort.
type PayloadExitSignal struct {
	Signal       string `json:"signal" yaml:"signal"`
	CoreDumped   bool   `json:"coreDumped" yaml:"coreDumped"`
	ErrorMessage string `json:"errorMessage" yaml:"errorMessage"`
	LanguageTag  string `json:"languageTag" yaml:"languageTag"`
}

// Equals compares two PayloadExitSignal payloads.
func (p PayloadExitSignal) Equals(other Payload) bool {
	p2, ok := other.(PayloadExitSignal)
	if !ok {
		return false
	}
	return p.Signal == p2.Signal &&
		p.CoreDumped == p2.CoreDumped &&
		p.ErrorMessage == p2.ErrorMessage &&
		p.LanguageTag == p2.LanguageTag
}
