package message

import "bytes"

// Stream The stream type corresponds to the file descriptor numbers common in UNIX systems for standard input, output,
//        and error.
type Stream uint

const (
	// StreamStdin Is the ID of the standard input that accepts testdata from the user.
	StreamStdin Stream = 0
	// StreamStdout Is the ID for the output stream containing normal messages or TTY-encoded testdata from the application.
	StreamStdout Stream = 1
	// StreamStderr Is the ID for the standard error containing the error messages for the application in non-TTY mode.
	StreamStderr Stream = 2
)

// PayloadIO The payload for I/O message types containing the testdata stream from/to the application.
type PayloadIO struct {
	Stream Stream `json:"stream" yaml:"stream"` // 0 = stdin, 1 = stdout, 2 = stderr
	Data   []byte `json:"data" yaml:"data"`
}

// Equals Compares two PayloadIO objects
func (p PayloadIO) Equals(other Payload) bool {
	p2, ok := other.(PayloadIO)
	if !ok {
		return false
	}
	return p.Stream == p2.Stream && bytes.Equal(p.Data, p2.Data)
}
