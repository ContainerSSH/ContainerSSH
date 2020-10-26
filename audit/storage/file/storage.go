package file

import (
	"io"
	"os"
	"path"
)

type Storage struct {
	directory string
}

func (s Storage) Open(name string) (io.WriteCloser, error) {
	return os.Create(path.Join(s.directory, name))
}
