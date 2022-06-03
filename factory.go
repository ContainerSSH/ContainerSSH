package libcontainerssh

import (
	"context"

	"go.containerssh.io/libcontainerssh/config"
	"go.containerssh.io/libcontainerssh/internal/auditlogintegration"
	"go.containerssh.io/libcontainerssh/internal/authintegration"
	"go.containerssh.io/libcontainerssh/internal/backend"
	"go.containerssh.io/libcontainerssh/internal/geoip"
	"go.containerssh.io/libcontainerssh/internal/geoip/geoipprovider"
	"go.containerssh.io/libcontainerssh/internal/health"
	"go.containerssh.io/libcontainerssh/internal/metrics"
	"go.containerssh.io/libcontainerssh/internal/metricsintegration"
	"go.containerssh.io/libcontainerssh/internal/sshserver"
	"go.containerssh.io/libcontainerssh/log"
	"go.containerssh.io/libcontainerssh/message"
	"go.containerssh.io/libcontainerssh/service"
)

// New creates a new instance of ContainerSSH.
func New(cfg config.AppConfig, factory log.LoggerFactory) (Service, service.Lifecycle, error) {
	if err := cfg.Validate(false); err != nil {
		return nil, nil, message.Wrap(err, message.ECoreConfig, "invalid ContainerSSH configuration")
	}

	logger, err := factory.Make(cfg.Log)
	if err != nil {
		return nil, nil, err
	}

	pool := service.NewPool(
		service.NewLifecycleFactory(),
		logger.WithLabel("module", "service"),
	)

	healthService, err := health.New(cfg.Health, logger.WithLabel("module", "health"))
	if err != nil {
		return nil, nil, err
	}
	pool.Add(healthService)

	geoIPLookupProvider, err := geoip.New(cfg.GeoIP)
	if err != nil {
		return nil, nil, err
	}

	metricsCollector := metrics.New(geoIPLookupProvider)

	if err := createMetricsServer(cfg, logger, metricsCollector, pool); err != nil {
		return nil, nil, err
	}

	containerBackend, err := createBackend(cfg, logger, metricsCollector)
	if err != nil {
		return nil, nil, err
	}

	authHandler, err := createAuthHandler(cfg, logger, containerBackend, metricsCollector, pool)
	if err != nil {
		return nil, nil, err
	}

	auditLogHandler, err := createAuditLogHandler(cfg, logger, authHandler, geoIPLookupProvider)
	if err != nil {
		return nil, nil, err
	}

	metricsHandler, err := createMetricsBackend(cfg, metricsCollector, auditLogHandler)
	if err != nil {
		return nil, nil, err
	}

	if err := createSSHServer(cfg, logger, metricsHandler, pool); err != nil {
		return nil, nil, err
	}

	return setUpService(pool, logger, healthService)
}

func setUpService(pool service.Pool, logger log.Logger, healthService health.Service) (
	Service,
	service.Lifecycle,
	error,
) {
	poolWrapper := &servicePool{
		pool,
		logger,
	}
	lifecycle := service.NewLifecycle(poolWrapper)
	lifecycle.OnRunning(
		func(s service.Service, l service.Lifecycle) {
			healthService.ChangeStatus(true)
		},
	).OnStopping(
		func(s service.Service, l service.Lifecycle, shutdownContext context.Context) {
			healthService.ChangeStatus(false)
		},
	).OnCrashed(
		func(s service.Service, l service.Lifecycle, err error) {
			healthService.ChangeStatus(false)
		},
	)

	return poolWrapper, lifecycle, nil
}

type servicePool struct {
	service.Pool
	logger log.Logger
}

func (s servicePool) RotateLogs() error {
	return s.logger.Rotate()
}

func createMetricsBackend(
	cfg config.AppConfig,
	collector metrics.Collector,
	handler sshserver.Handler,
) (sshserver.Handler, error) {
	return metricsintegration.NewHandler(
		cfg.Metrics,
		collector,
		handler,
	)
}

func createMetricsServer(
	cfg config.AppConfig,
	logger log.Logger,
	metricsCollector metrics.Collector,
	pool service.Pool,
) error {
	metricsLogger := logger.WithLabel("module", "metrics")
	metricsServer, err := metrics.NewServer(cfg.Metrics, metricsCollector, metricsLogger)
	if err != nil {
		return err
	}
	if metricsServer == nil {
		return nil
	}
	pool.Add(metricsServer)
	return nil
}

func createSSHServer(
	cfg config.AppConfig,
	logger log.Logger,
	auditLogHandler sshserver.Handler,
	pool service.Pool,
) error {
	sshLogger := logger.WithLabel("module", "ssh")
	sshServer, err := sshserver.New(
		cfg.SSH,
		auditLogHandler,
		sshLogger,
	)
	if err != nil {
		return err
	}
	pool.Add(sshServer)
	return nil
}

func createAuditLogHandler(
	cfg config.AppConfig,
	logger log.Logger,
	authHandler sshserver.Handler,
	geoIPLookupProvider geoipprovider.LookupProvider,
) (sshserver.Handler, error) {
	auditLogger := logger.WithLabel("module", "audit")
	return auditlogintegration.New(
		cfg.Audit,
		authHandler,
		geoIPLookupProvider,
		auditLogger,
	)
}

func createAuthHandler(
	cfg config.AppConfig,
	logger log.Logger,
	backend sshserver.Handler,
	metricsCollector metrics.Collector,
	pool service.Pool,
) (sshserver.Handler, error) {
	authLogger := logger.WithLabel("module", "auth")
	handler, services, err := authintegration.New(
		cfg.Auth,
		backend,
		authLogger,
		metricsCollector,
		authintegration.BehaviorNoPassthrough,
	)
	if err != nil {
		return nil, err
	}
	for _, svc := range services {
		pool.Add(svc)
	}
	return handler, nil
}

func createBackend(cfg config.AppConfig, logger log.Logger, metricsCollector metrics.Collector) (sshserver.Handler, error) {
	backendLogger := logger.WithLabel("module", "backend")
	containerBackend, err := backend.New(cfg, backendLogger, metricsCollector, sshserver.AuthResponseUnavailable)
	if err != nil {
		return nil, err
	}
	return containerBackend, nil
}
