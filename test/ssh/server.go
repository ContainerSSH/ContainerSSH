package ssh

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/janoszen/containerssh/auth"
	"github.com/janoszen/containerssh/backend"
	"github.com/janoszen/containerssh/backend/dockerrun"
	"github.com/janoszen/containerssh/backend/kuberun"
	configurationClient "github.com/janoszen/containerssh/config/client"
	"github.com/janoszen/containerssh/config/util"
	"github.com/janoszen/containerssh/log"
	"github.com/janoszen/containerssh/ssh"
)

type Server struct {
	ctx context.Context
	cancel context.CancelFunc
	logger log.Logger
	logWriter log.Writer
	authClient auth.Client
	configClient configurationClient.ConfigClient
}

func NewServer(
	logger log.Logger,
	logWriter log.Writer,
	authClient auth.Client,
	configClient configurationClient.ConfigClient,
) *Server {
	return &Server {
		logWriter: logWriter,
		logger: logger,
		authClient: authClient,
		configClient: configClient,
	}
}

func (server *Server) Start() error {
	if server.ctx == nil {
		server.ctx, server.cancel = context.WithCancel(context.Background())

		backendRegistry := backend.NewRegistry()
		dockerrun.Init(backendRegistry)
		kuberun.Init(backendRegistry)

		appConfig, err := util.GetDefaultConfig()

		privateKey, err := rsa.GenerateKey(rand.Reader, 2014)
		if err != nil {
			return err
		}
		privateKeyDer := x509.MarshalPKCS1PrivateKey(privateKey)
		privateKeyBlock := pem.Block{
			Type:    "RSA PRIVATE KEY",
			Headers: nil,
			Bytes:   privateKeyDer,
		}
		privateKeyPem := string(pem.EncodeToMemory(&privateKeyBlock))
		appConfig.Ssh.HostKeys = append(appConfig.Ssh.HostKeys, privateKeyPem)

		sshServer, err := ssh.NewServer(
			appConfig,
			server.authClient,
			backendRegistry,
			server.configClient,
			server.logger,
			server.logWriter,
		)
		if err != nil {
			server.logger.EmergencyF("failed to create SSH server (%v)", err)
			return err
		}

		go func() {
			err = sshServer.Run(server.ctx)
			if err != nil {
				server.logger.EmergencyF("failed to start SSH server (%v)", err)
				server.cancel()
			}
		}()
	}
	return nil
}

func (server *Server) Stop() error {
	if server.cancel == nil {
		return fmt.Errorf("server is not running")
	}
	server.cancel()
	server.cancel = nil
	server.ctx = nil
	return nil
}