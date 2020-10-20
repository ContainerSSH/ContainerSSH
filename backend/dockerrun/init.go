package dockerrun

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"github.com/docker/docker/client"
	"github.com/janoszen/containerssh/backend"
	"github.com/janoszen/containerssh/config"
	"github.com/janoszen/containerssh/log"
	"github.com/janoszen/containerssh/metrics"
	"net/http"
)

func createSession(sessionId string, username string, appConfig *config.AppConfig, logger log.Logger, metric *metrics.MetricCollector) (backend.Session, error) {
	logger.DebugF("initializing Docker backend")
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
	session.logger = logger
	session.metric = metric

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
	logger      log.Logger
	metric      *metrics.MetricCollector
}

func Init(registry *backend.Registry, metric *metrics.MetricCollector) {
	metric.SetMetricMeta(MetricNameBackendError, "Number of errors in the dockerrun backend", metrics.MetricTypeCounter)
	metric.Set(MetricBackendError, 0)

	dockerRunBackend := backend.Backend{}
	dockerRunBackend.Name = "dockerrun"
	dockerRunBackend.CreateSession = createSession
	registry.Register(dockerRunBackend)
}
