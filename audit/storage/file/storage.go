package file

import (
	"github.com/containerssh/containerssh/audit"
	"os"
	"path"
)

type Storage struct {
	directory string
}

func (s Storage) Open(name string) (audit.StorageWriter, error) {
	return audit.NewStorageWriterProxy(os.Create(path.Join(s.directory, name)))
}
