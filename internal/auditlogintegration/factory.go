package auditlogintegration

import (
	"fmt"

	"github.com/containerssh/libcontainerssh/config"
	"github.com/containerssh/libcontainerssh/internal/auditlog"
	"github.com/containerssh/libcontainerssh/internal/geoip/geoipprovider"
	"github.com/containerssh/libcontainerssh/internal/sshserver"
	"github.com/containerssh/libcontainerssh/log"
)

// New creates a new handler based on the application config and the required dependencies. If audit logging is not
// enabled the backend will be returned directly.
//goland:noinspection GoUnusedExportedFunction
func New(
	cfg config.AuditLogConfig,
	backend sshserver.Handler,
	geoIPLookupProvider geoipprovider.LookupProvider,
	logger log.Logger,
) (sshserver.Handler, error) {
	if !cfg.Enable {
		return backend, nil
	}

	auditLogger, err := auditlog.New(
		cfg,
		geoIPLookupProvider,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create audit logger (%w)", err)
	}

	handler := NewHandler(
		backend,
		auditLogger,
	)
	return handler, nil
}
