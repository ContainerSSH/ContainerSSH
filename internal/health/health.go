package health

import (
	"fmt"

	"go.containerssh.io/libcontainerssh/config"
	http2 "go.containerssh.io/libcontainerssh/http"
	"go.containerssh.io/libcontainerssh/log"
	"go.containerssh.io/libcontainerssh/message"
	"go.containerssh.io/libcontainerssh/service"
)

// New creates a new HTTP health service on port 23074
func New(cfg config.HealthConfig, logger log.Logger) (Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	handler := &requestHandler{}
	svc, err := http2.NewServer(
		"Health check endpoint",
		cfg.HTTPServerConfiguration,
		http2.NewServerHandlerNegotiate(handler, logger),
		logger,
		func(url string) {
			logger.Info(message.NewMessage(message.MHealthServiceAvailable, "Health check endpoint available at %s", url))
		},
	)
	if err != nil {
		return nil, err
	}

	return &healthCheckService{
		Service:        svc,
		requestHandler: handler,
	}, nil
}

// NewClient creates a new health check client based on the supplied configuration. If the health check is not enabled
// no client is returned.
func NewClient(cfg config.HealthConfig, logger log.Logger) (Client, error) {
	if !cfg.Enable {
		return nil, nil
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid health check configuration (%w)", err)
	}

	httpClient, err := http2.NewClient(cfg.Client, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create health check client (%w)", err)
	}

	return &healthCheckClient{
		httpClient: httpClient,
		logger:     logger,
	}, nil
}

// Service is an HTTP service that lets you change the status via ChangeStatus().
type Service interface {
	service.Service
	ChangeStatus(ok bool)
}

// Client is the client to run health checks.
type Client interface {
	// Run runs an HTTP query against the health check service.
	Run() bool
}

type healthCheckService struct {
	service.Service
	requestHandler *requestHandler
}

func (h *healthCheckService) ChangeStatus(ok bool) {
	h.requestHandler.ok = ok
}

type requestHandler struct {
	ok bool
}

func (r requestHandler) OnRequest(_ http2.ServerRequest, response http2.ServerResponse) error {
	if r.ok {
		response.SetBody("ok")
	} else {
		response.SetBody("not ok")
		response.SetStatus(503)
	}
	return nil
}

type healthCheckClient struct {
	httpClient http2.Client
	logger     log.Logger
}

func (h *healthCheckClient) Run() bool {
	responseBody := ""
	statusCode, err := h.httpClient.Get("", &responseBody)
	if err != nil {
		h.logger.Warning(message.Wrap(err, message.EHealthRequestFailed, "Request to health check endpoint failed"))
	}
	return statusCode == 200
}
