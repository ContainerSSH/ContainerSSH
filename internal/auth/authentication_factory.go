package auth

import (
	"fmt"

	"go.containerssh.io/containerssh/config"
	"go.containerssh.io/containerssh/internal/metrics"
	"go.containerssh.io/containerssh/log"
	"go.containerssh.io/containerssh/message"
	"go.containerssh.io/containerssh/service"
)

// NewPasswordAuthenticator returns a password authenticator as configured, and if needed a backing service that needs
// to run for the authentication to work. If password authentication is disabled it returns nil. If the configuration is
// invalid an error is returned.
func NewPasswordAuthenticator(
	cfg config.PasswordAuthConfig,
	logger log.Logger,
	metrics metrics.Collector,
) (PasswordAuthenticator, service.Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, nil, err
	}
	switch cfg.Method {
	case config.PasswordAuthMethodDisabled:
		return nil, nil, nil
	case config.PasswordAuthMethodWebhook:
		cli, err := NewWebhookClient(AuthenticationTypePassword, cfg.Webhook, logger, metrics)
		return cli, nil, err
	case config.PasswordAuthMethodKerberos:
		cli, err := NewKerberosClient(AuthenticationTypePassword, cfg.Kerberos, logger, metrics)
		return cli, nil, err
	default:
		return nil, nil, fmt.Errorf("unsupported method: %s", cfg.Method)
	}
}

// NewPublicKeyAuthenticator returns a public key authenticator as configured, and if needed a backing service that
// needs to run for the authentication to work. If public key authentication is disabled it returns nil. If the
// configuration is invalid an error is returned.
func NewPublicKeyAuthenticator(
	cfg config.PublicKeyAuthConfig,
	logger log.Logger,
	metrics metrics.Collector,
) (PublicKeyAuthenticator, service.Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, nil, err
	}
	switch cfg.Method {
	case config.PubKeyAuthMethodDisabled:
		return nil, nil, nil
	case config.PubKeyAuthMethodWebhook:
		cli, err := NewWebhookClient(AuthenticationTypePublicKey, cfg.Webhook, logger, metrics)
		return cli, nil, err
	default:
		return nil, nil, fmt.Errorf("unsupported method: %s", cfg.Method)
	}
}

// NewKeyboardInteractiveAuthenticator returns a keyboard-interactive authenticator as configured, and if needed a
// backing service that needs to run for the authentication to work. If keyboard-interactive authentication is disabled
// it returns nil. If the configuration is invalid an error is returned.
func NewKeyboardInteractiveAuthenticator(
	cfg config.KeyboardInteractiveAuthConfig,
	logger log.Logger,
	metrics metrics.Collector,
) (KeyboardInteractiveAuthenticator, service.Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, nil, err
	}
	switch cfg.Method {
	case config.KeyboardInteractiveAuthMethodDisabled:
		return nil, nil, nil
	case config.KeyboardInteractiveAuthMethodOAuth2:
		return NewOAuth2Client(cfg.OAuth2, logger, metrics)
	default:
		return nil, nil, fmt.Errorf("unsupported method: %s", cfg.Method)
	}
}

// NewNoneAuthenticator returns a none authenticator as configured. If none authentication is disabled it returns
// a noneAuthenticator that always returns failure. If the configuration is invalid an error is returned.
func NewNoneAuthenticator(
	cfg config.NoneAuthConfig,
	logger log.Logger,
	metrics metrics.Collector,
) (NoneAuthenticator, error) {
	if cfg.Enabled {
		logger.Warning(message.NewMessage(
			message.WNoneAuthEnabled,
			"None authentication is enabled. This allows any user to log in without a password. This is insecure and should only be used in secure environments.",
		))
	}
	return &noneAuthenticator{
		enabled: cfg.Enabled,
	}, nil
}

// NewGSSAPIAuthenticator returns a GSSAPI authenticator as configured, and if needed a backing service that needs to
// run for the authentication to work. If GSSAPI authentication is disabled it returns nil. If the configuration is
// invalid an error is returned.
func NewGSSAPIAuthenticator(
	cfg config.GSSAPIAuthConfig,
	logger log.Logger,
	metrics metrics.Collector,
) (GSSAPIAuthenticator, service.Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, nil, err
	}
	switch cfg.Method {
	case config.GSSAPIAuthMethodDisabled:
		return nil, nil, nil
	case config.GSSAPIAuthMethodKerberos:
		cli, err := NewKerberosClient(AuthenticationTypeGSSAPI, cfg.Kerberos, logger, metrics)
		return cli, nil, err
	default:
		return nil, nil, fmt.Errorf("unsupported method: %s", cfg.Method)
	}
}

// NewAuthorizationProvider returns an authorization provider as configured, and if needed a backing service that needs to
// run for the authorization to work. If authorization is disabled it returns nil. If the configuration is
// invalid an error is returned.
func NewAuthorizationProvider(
	cfg config.AuthzConfig,
	logger log.Logger,
	metrics metrics.Collector,
) (AuthzProvider, service.Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, nil, err
	}
	switch cfg.Method {
	case config.AuthzMethodDisabled:
		return nil, nil, nil
	case config.AuthzMethodWebhook:
		cli, err := NewWebhookClient(AuthenticationTypeAuthz, cfg.Webhook, logger, metrics)
		return cli, nil, err
	default:
		return nil, nil, fmt.Errorf("unsupported method: %s", cfg.Method)
	}
}
