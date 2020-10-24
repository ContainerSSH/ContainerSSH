package file

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"

	"github.com/containerssh/containerssh/audit"
	"github.com/containerssh/containerssh/config"
	"github.com/containerssh/containerssh/log"
)

func New(cfg config.AuditFileConfig, logger log.Logger) (audit.Plugin, error) {
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
	return &Plugin{
		directory:   cfg.Directory,
		connections: sync.Map{},
		logger:      logger,
	}, nil
}
