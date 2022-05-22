package authintegration

import (
	"fmt"

	"github.com/containerssh/libcontainerssh/config"
	"github.com/containerssh/libcontainerssh/internal/auth"
	"github.com/containerssh/libcontainerssh/internal/metrics"
	"github.com/containerssh/libcontainerssh/internal/sshserver"
	"github.com/containerssh/libcontainerssh/log"
	"github.com/containerssh/libcontainerssh/service"
)

// New creates a new handler that authenticates the users with passwords and public keys.
//goland:noinspection GoUnusedExportedFunction
func New(
	config config.AuthConfig,
	backend sshserver.Handler,
	logger log.Logger,
	metricsCollector metrics.Collector,
	behavior Behavior,
) (sshserver.Handler, []service.Service, error) {
	if backend == nil {
		return nil, nil, fmt.Errorf("the backend parameter to authintegration.New cannot be nil")
	}
	if !behavior.validate() {
		return nil, nil, fmt.Errorf("the behavior field contains an invalid value: %d", behavior)
	}

	var services []service.Service

	passwordAuthenticator, svc, err := auth.NewPasswordAuthenticator(config.PasswordAuth, logger, metricsCollector)
	if err != nil {
		return nil, nil, err
	}
	if svc != nil {
		services = append(services, svc)
	}

	publicKeyAuthenticator, svc, err := auth.NewPublicKeyAuthenticator(config.PublicKeyAuth, logger, metricsCollector)
	if err != nil {
		return nil, nil, err
	}
	if svc != nil {
		services = append(services, svc)
	}

	keyboardInteractiveAuthenticator, svc, err := auth.NewKeyboardInteractiveAuthenticator(
		config.KeyboardInteractiveAuth,
		logger,
		metricsCollector,
	)
	if err != nil {
		return nil, nil, err
	}
	if svc != nil {
		services = append(services, svc)
	}

	gssapiAuthenticator, svc, err := auth.NewGSSAPIAuthenticator(
		config.GSSAPIAuth,
		logger,
		metricsCollector,
	)
	if err != nil {
		return nil, nil, err
	}
	if svc != nil {
		services = append(services, svc)
	}

	authorizationProvider, svc, err := auth.NewAuthorizationProvider(config.Authz, logger, metricsCollector)
	if err != nil {
		return nil, nil, err
	}
	if svc != nil {
		services = append(services, svc)
	}

	return &handler{
		passwordAuthenticator:            passwordAuthenticator,
		publicKeyAuthenticator:           publicKeyAuthenticator,
		keyboardInteractiveAuthenticator: keyboardInteractiveAuthenticator,
		gssapiAuthenticator:              gssapiAuthenticator,
		authorizationProvider:            authorizationProvider,
		backend:                          backend,
		behavior:                         behavior,
	}, services, nil
}
