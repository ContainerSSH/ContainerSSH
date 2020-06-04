package ssh

const (
	REQ_SHELL         = "shell"
	REQ_SUBSYSTEM     = "subsystem"
	REQ_PTY           = "pty-req"
	REQ_WINDOW_CHANGE = "window-change"
	REQ_SET_ENV       = "env"
	REQ_SIGNAL        = "signal"
	REQ_EXIT_STATUS   = "exit-status"
)

type PtyRequestMsg struct {
	Term     string
	Columns  uint32
	Rows     uint32
	Width    uint32
	Height   uint32
	Modelist string
}

type PtyWindowChangeMsg struct {
	Columns uint32
	Rows    uint32
	Width   uint32
	Height  uint32
}

type SetenvRequest struct {
	Name  string
	Value string
}

type SubsystemRequestMsg struct {
	Subsystem string
}

type SignalMsg struct {
	Signal string
}

type ExitStatusMsg struct {
	ExitStatus uint32
}

type ExitSignalMsg struct {
	Signal       string
	CoreDumped   bool
	ErrorMessage string
	LanguageTag  string
}
