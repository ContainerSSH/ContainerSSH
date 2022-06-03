package http

import (
	"crypto/tls"
	"strings"

    "go.containerssh.io/libcontainerssh/config"
    "go.containerssh.io/libcontainerssh/log"
)

// NewClient creates a new HTTP client with the given configuration.
func NewClient(
	config config.HTTPClientConfiguration,
	logger log.Logger,
) (Client, error) {
	return NewClientWithHeaders(config, logger, nil, false)
}

// NewClientWithHeaders creates a new HTTP client with added extra headers.
func NewClientWithHeaders(
	config config.HTTPClientConfiguration,
	logger log.Logger,
	extraHeaders map[string][]string,
	allowLaxDecoding bool,
) (Client, error) {
	certs, err := config.ValidateWithCerts()
	if err != nil {
		return nil, err
	}
	if logger == nil {
		panic("BUG: no logger provided for http.NewClient")
	}

	tlsConfig := createTLSConfig(config, certs)

	return &client{
		config:           config,
		logger:           logger.WithLabel("endpoint", config.URL),
		tlsConfig:        tlsConfig,
		extraHeaders:     extraHeaders,
		allowLaxDecoding: allowLaxDecoding,
	}, nil
}

// createTLSConfig creates a TLS config. Should only be called after config.Validate().
func createTLSConfig(config config.HTTPClientConfiguration, certs *config.HTTPClientCerts) *tls.Config {
	if !strings.HasPrefix(config.URL, "https://") {
		return nil
	}

	// We let users configure the minimum TLS version, so we don't need gosec here.
	tlsConfig := &tls.Config{ //nolint:gosec
		MinVersion:       config.TLSVersion.GetTLSVersion(),
		CurvePreferences: config.ECDHCurves.GetList(),
		CipherSuites:     config.CipherSuites.GetList(),
	}
	if certs.CACertPool != nil {
		tlsConfig.RootCAs = certs.CACertPool
	}
	if certs.Cert != nil {
		tlsConfig.Certificates = []tls.Certificate{*certs.Cert}
	}
	return tlsConfig
}
