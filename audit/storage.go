package audit

import "io"

type Storage interface {
	Open(name string) (io.WriteCloser, error)
}
