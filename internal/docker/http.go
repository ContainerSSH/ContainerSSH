package docker

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"strings"

	"github.com/containerssh/containerssh/config"
)

func getHTTPClient(config config.DockerConfig) (*http.Client, error) {
	var httpClient *http.Client = nil
	if config.Connection.CaCert != "" && config.Connection.Key != "" && config.Connection.Cert != "" {
		tlsConfig := &tls.Config{}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM([]byte(config.Connection.CaCert))
		tlsConfig.RootCAs = caCertPool

		keyPair, err := tls.X509KeyPair([]byte(config.Connection.Cert), []byte(config.Connection.Key))
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{keyPair}
		transport := &http.Transport{TLSClientConfig: tlsConfig}
		httpClient = &http.Client{
			Transport: transport,
			Timeout:   config.Timeouts.HTTP,
		}
	} else if strings.HasPrefix(config.Connection.Host, "http://") {
		httpClient = &http.Client{
			Transport: http.DefaultTransport,
			Timeout:   config.Timeouts.HTTP,
		}
	}
	return httpClient, nil
}
