package dockerrun

import (
	"containerssh/backend"
	"containerssh/config"
	"context"
	"github.com/docker/docker/client"
)

func createSession(sessionId string, username string, appConfig *config.AppConfig) (backend.Session, error) {
	cli, err := client.NewClient("tcp://127.0.0.1:2375", "", nil, make(map[string]string))
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
