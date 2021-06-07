package containerssh

import (
	"github.com/containerssh/health"
	"github.com/containerssh/log"
	"github.com/containerssh/service"
	"github.com/containerssh/sshserver"

	"github.com/containerssh/auditlogintegration"
	"github.com/containerssh/authintegration"
	"github.com/containerssh/backend/v2"
	"github.com/containerssh/configuration/v2"
	"github.com/containerssh/geoip"
	"github.com/containerssh/geoip/geoipprovider"
	"github.com/containerssh/metrics"
	"github.com/containerssh/metricsintegration"
)

// New creates a new instance of ContainerSSH.
func New(config configuration.AppConfig, factory log.LoggerFactory) (Service, error) {
	if err := config.Validate(false); err != nil {
		return nil, log.Wrap(err, EConfig, "invalid ContainerSSH configuration")
	}

	logger, err := factory.Make(config.Log)
	if err != nil {
		return nil, err
	}

	pool := service.NewPool(
		service.NewLifecycleFactory(),
		logger.WithLabel("module", "service"),
	)

	healthService, err := health.New(config.Health, logger.WithLabel("module", "health"))
	if err != nil {
		return nil, err
	}
	pool.Add(healthService)

	geoIPLookupProvider, err := geoip.New(config.GeoIP)
	if err != nil {
		return nil, err
	}

	metricsCollector := metrics.New(geoIPLookupProvider)

	if err := createMetricsServer(config, logger, metricsCollector, pool); err != nil {
		return nil, err
	}

	containerBackend, err := createBackend(config, logger, metricsCollector)
	if err != nil {
		return nil, err
	}

	authHandler, err := createAuthHandler(config, logger, containerBackend, metricsCollector)
	if err != nil {
		return nil, err
	}

	auditLogHandler, err := createAuditLogHandler(config, logger, authHandler, geoIPLookupProvider)
	if err != nil {
		return nil, err
	}

	metricsHandler, err := createMetricsBackend(config, metricsCollector, auditLogHandler)
	if err != nil {
		return nil, err
	}

	if err := createSSHServer(config, logger, metricsHandler, pool); err != nil {
		return nil, err
	}

	return &servicePool{
		pool,
		logger,
	}, nil
}

type servicePool struct {
	service.Pool
	logger log.Logger
}

func (s servicePool) RotateLogs() error {
	return s.logger.Rotate()
}

func createMetricsBackend(
	config configuration.AppConfig,
	collector metrics.Collector,
	handler sshserver.Handler,
) (sshserver.Handler, error) {
	return metricsintegration.NewHandler(
		config.Metrics,
		collector,
		handler,
	)
}

func createMetricsServer(
	config configuration.AppConfig,
	logger log.Logger,
	metricsCollector metrics.Collector,
	pool service.Pool,
) error {
	metricsLogger := logger.WithLabel("module", "metrics")
	metricsServer, err := metrics.NewServer(config.Metrics, metricsCollector, metricsLogger)
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
	config configuration.AppConfig,
	logger log.Logger,
	auditLogHandler sshserver.Handler,
	pool service.Pool,
) error {
	sshLogger := logger.WithLabel("module", "ssh")
	sshServer, err := sshserver.New(
		config.SSH,
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
	config configuration.AppConfig,
	logger log.Logger,
	authHandler sshserver.Handler,
	geoIPLookupProvider geoipprovider.LookupProvider,
) (sshserver.Handler, error) {
	auditLogger := logger.WithLabel("module", "audit")
	return auditlogintegration.New(
		config.Audit,
		authHandler,
		geoIPLookupProvider,
		auditLogger,
	)
}

func createAuthHandler(
	config configuration.AppConfig,
	logger log.Logger,
	backend sshserver.Handler,
	metricsCollector metrics.Collector,
) (sshserver.Handler, error) {
	authLogger := logger.WithLabel("module", "auth")
	return authintegration.New(
		config.Auth,
		backend,
		authLogger,
		metricsCollector,
		authintegration.BehaviorNoPassthrough,
	)
}

func createBackend(config configuration.AppConfig, logger log.Logger, metricsCollector metrics.Collector) (sshserver.Handler, error) {
	backendLogger := logger.WithLabel("module", "backend")
	containerBackend, err := backend.New(config, backendLogger, metricsCollector, sshserver.AuthResponseUnavailable)
	if err != nil {
		return nil, err
	}
	return containerBackend, nil
}
