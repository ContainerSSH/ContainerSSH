package file

import (
	"fmt"
	"os"
	"path"
	"sync"

	"go.containerssh.io/libcontainerssh/config"
	"go.containerssh.io/libcontainerssh/internal/auditlog/storage"

	"go.containerssh.io/libcontainerssh/log"
)

// NewStorage Create a file storage that stores testdata in a local directory. The file storage cannot store metadata.
func NewStorage(cfg config.AuditLogFileConfig, _ log.Logger) (storage.ReadWriteStorage, error) {
	if cfg.Directory == "" {
		return nil, fmt.Errorf("invalid audit log directory")
	}
	stat, err := os.Stat(cfg.Directory)
	if err != nil {
		return nil, fmt.Errorf("failed to access audit log directory %s (%w)", cfg.Directory, err)
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf("specified audit log directory is not a directory %s (%w)", cfg.Directory, err)
	}
	err = os.WriteFile(path.Join(cfg.Directory, ".accesstest"), []byte{}, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to create file in audit log directory %s (%w)", cfg.Directory, err)
	}
	return &fileStorage{
		directory: cfg.Directory,
		wg:        &sync.WaitGroup{},
	}, nil
}
