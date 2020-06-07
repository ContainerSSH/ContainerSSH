package dockerrun

import (
	"containerssh/backend"
	"containerssh/config"
	"context"
	"crypto/tls"
	"crypto/x509"
	"github.com/docker/docker/client"
	"net/http"
)

func createSession(sessionId string, username string, appConfig *config.AppConfig) (backend.Session, error) {
	var httpClient *http.Client = nil
	if appConfig.DockerRun.CaCert != "" && appConfig.DockerRun.Key != "" && appConfig.DockerRun.Cert != "" {
		tlsConfig := &tls.Config{}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM([]byte(appConfig.DockerRun.CaCert))
		tlsConfig.RootCAs = caCertPool

		keyPair, err := tls.X509KeyPair([]byte(appConfig.DockerRun.Cert), []byte(appConfig.DockerRun.Key))
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{keyPair}
		transport := &http.Transport{TLSClientConfig: tlsConfig}
		httpClient = &http.Client{
			Transport: transport,
		}
	}

	cli, err := client.NewClient(appConfig.DockerRun.Host, "", httpClient, make(map[string]string))
	if err != nil {
		return nil, err
	}

	session := &dockerRunSession{}
	session.sessionId = sessionId
	session.username = username
	session.env = map[string]string{}
	session.cols = 80
	session.rows = 25
	session.pty = false
	session.containerId = ""
	session.client = cli
	session.ctx = context.Background()
	session.exitCode = -1
	session.config = &appConfig.DockerRun

	return session, nil
}

type dockerRunSession struct {
	username    string
	sessionId   string
	env         map[string]string
	cols        uint
	rows        uint
	width       uint
	height      uint
	pty         bool
	containerId string
	exitCode    int32
	ctx         context.Context
	client      *client.Client
	config      *config.DockerRunConfig
}

func Init(registry *backend.Registry) {
	dockerRunBackend := backend.Backend{}
	dockerRunBackend.Name = "dockerrun"
	dockerRunBackend.CreateSession = createSession
	registry.Register(dockerRunBackend)
}
