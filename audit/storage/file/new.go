package file

import (
	"fmt"
	"github.com/containerssh/containerssh/audit"
	"github.com/containerssh/containerssh/config"
	"github.com/containerssh/containerssh/log"
	"io/ioutil"
	"os"
	"path"
)

func NewStorage(cfg config.AuditFileConfig, _ log.Logger) (audit.Storage, error) {
	if cfg.Directory == "" {
		return nil, fmt.Errorf("invalid audit log directory")
	}
	stat, err := os.Stat(cfg.Directory)
	if err != nil {
		return nil, fmt.Errorf("failed to access audit log directory %s (%v)", cfg.Directory, err)
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf("specified audit log directory is not a directory %s (%v)", cfg.Directory, err)
	}
	err = ioutil.WriteFile(path.Join(cfg.Directory, ".accesstest"), []byte{}, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create file in audit log directory %s (%v)", cfg.Directory, err)
	}
	return &Storage{
		directory: cfg.Directory,
	}, nil
}
