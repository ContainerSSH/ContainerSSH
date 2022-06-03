package http

import (
	"crypto/tls"
	goHttp "net/http"
	"sync"

    "go.containerssh.io/libcontainerssh/config"
    "go.containerssh.io/libcontainerssh/log"
)

// NewServer creates a new HTTP server with the given configuration and calling the provided handler.
func NewServer(
	name string,
	config config.HTTPServerConfiguration,
	handler goHttp.Handler,
	logger log.Logger,
	onReady func(string),
) (Server, error) {
	if handler == nil {
		panic("BUG: no handler provided to http.NewServer")
	}
	if logger == nil {
		panic("BUG: no logger provided to http.NewServer")
	}

	certs, err := config.ValidateWithCerts()
	if err != nil {
		return nil, err
	}

	var tlsConfig *tls.Config
	if certs.Cert != nil {
		tlsConfig = createServerTLSConfig(config, certs)
	}

	return &server{
		name:      name,
		lock:      &sync.Mutex{},
		handler:   handler,
		config:    config,
		tlsConfig: tlsConfig,
		srv:       nil,
		goLogger:  log.NewGoLogWriter(logger),
		onReady:   onReady,
	}, nil
}

func createServerTLSConfig(config config.HTTPServerConfiguration, certs *config.HTTPServerCerts) *tls.Config {
	// We let users configure the minimum TLS version, so we don't need gosec here.
	tlsConfig := &tls.Config{ //nolint:gosec
		MinVersion:               config.TLSVersion.GetTLSVersion(),
		CurvePreferences:         config.ECDHCurves.GetList(),
		PreferServerCipherSuites: true,
		CipherSuites:             config.CipherSuites.GetList(),
	}

	tlsConfig.Certificates = []tls.Certificate{*certs.Cert}

	if certs.ClientCAPool != nil {
		tlsConfig.ClientCAs = certs.ClientCAPool
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}
	return tlsConfig
}
