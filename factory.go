package containerssh

import (
	"context"

	"github.com/containerssh/auditlogintegration"
	"github.com/containerssh/authintegration"
	"github.com/containerssh/backend"
	"github.com/containerssh/configuration"
	"github.com/containerssh/geoip"
	"github.com/containerssh/geoip/geoipprovider"
	"github.com/containerssh/log"
	"github.com/containerssh/service"
	"github.com/containerssh/sshserver"
)

func New(config configuration.AppConfig, factory log.LoggerFactory) (service.Service, error) {
	pool := service.NewPool(
		service.NewLifecycleFactory(),
	)

	containerBackend, err := createBackend(config, factory)
	if err != nil {
		return nil, err
	}

	authHandler, err := createAuthHandler(config, factory, containerBackend)
	if err != nil {
		return nil, err
	}

	geoIPLookupProvider, err := geoip.New(config.GeoIP)
	if err != nil {
		return nil, err
	}

	auditLogHandler, err := createAuditLogHandler(config, factory, authHandler, geoIPLookupProvider)
	if err != nil {
		return nil, err
	}

	if err := createSSHServer(config, factory, auditLogHandler, pool); err != nil {
		return nil, err
	}

	return pool, nil
}

func createSSHServer(
	config configuration.AppConfig,
	factory log.LoggerFactory,
	auditLogHandler sshserver.Handler,
	pool service.Pool,
) error {
	sshLogger, err := factory.Make(config.Log, "ssh")
	if err != nil {
		return err
	}
	sshServer, err := sshserver.New(
		config.SSH,
		auditLogHandler,
		sshLogger,
	)
	if err != nil {
		return err
	}
	pool.Add(sshServer).OnStarting(
		func(s service.Service, l service.Lifecycle) {
			sshLogger.Noticef("SSH service is starting...")
		},
	).OnRunning(
		func(s service.Service, l service.Lifecycle) {
			sshLogger.Noticef("SSH service is running at %s", config.SSH.Listen)
		},
	).OnStopping(
		func(s service.Service, l service.Lifecycle, shutdownContext context.Context) {
			sshLogger.Noticef("SSH service is stopping...")
		},
	).OnStopped(
		func(s service.Service, l service.Lifecycle) {
			sshLogger.Noticef("SSH service has stopped.")
		},
	)
	return nil
}

func createAuditLogHandler(
	config configuration.AppConfig,
	factory log.LoggerFactory,
	authHandler sshserver.Handler,
	geoIPLookupProvider geoipprovider.LookupProvider,
) (sshserver.Handler, error) {
	auditLogger, err := factory.Make(config.Log, "audit")
	if err != nil {
		return nil, err
	}
	return auditlogintegration.New(
		config.Audit,
		authHandler,
		geoIPLookupProvider,
		auditLogger,
	)
}

func createAuthHandler(
	config configuration.AppConfig,
	factory log.LoggerFactory,
	backend sshserver.Handler,
) (sshserver.Handler, error) {
	authLogger, err := factory.Make(config.Log, "auth")
	if err != nil {
		return nil, err
	}
	return authintegration.New(
		config.Auth,
		backend,
		authLogger,
		authintegration.BehaviorNoPassthrough,
	)
}

func createBackend(config configuration.AppConfig, factory log.LoggerFactory) (sshserver.Handler, error) {
	backendLogger, err := factory.Make(config.Log, "backend")
	if err != nil {
		return nil, err
	}
	containerBackend, err := backend.New(config, backendLogger, factory, sshserver.AuthResponseUnavailable)
	if err != nil {
		return nil, err
	}
	return containerBackend, nil
}
