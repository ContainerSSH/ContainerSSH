package audit

type Stream uint

const (
	Stream_Stdin  Stream = 0
	Stream_StdOut Stream = 1
	Stream_StdErr Stream = 2
)

type PayloadIO struct {
	Stream Stream `json:"stream" yaml:"stream"`
	Data   []byte `json:"data" yaml:"data"`
}
