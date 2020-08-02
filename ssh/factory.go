package ssh

import (
	"fmt"
	"github.com/janoszen/containerssh/auth"
	"github.com/janoszen/containerssh/backend"
	"github.com/janoszen/containerssh/config"
	configurationClient "github.com/janoszen/containerssh/config/client"
	"github.com/janoszen/containerssh/log"
	"github.com/janoszen/containerssh/ssh/server"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"strings"
)

func NewServer(
	config *config.AppConfig,
	authClient auth.Client,
	registry *backend.Registry,
	client configurationClient.ConfigClient,
	logger log.Logger,
	logWriter log.Writer,
) (*server.Server, error) {
	serverConfig := &server.Config{
		Config:       ssh.Config{},
		NoClientAuth: false,
		MaxAuthTries: 6,
	}

	if config.Auth.Password {
		serverConfig.PasswordCallback = func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			authResponse, err := authClient.Password(
				conn.User(),
				password,
				conn.SessionID(),
				conn.RemoteAddr().String(),
			)
			if err != nil {
				return nil, err
			}
			if !authResponse.Success {
				return nil, fmt.Errorf("authentication failed")
			}
			return &ssh.Permissions{}, nil
		}
	}

	if config.Auth.PubKey {
		serverConfig.PublicKeyCallback = func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			authResponse, err := authClient.PubKey(
				conn.User(),
				key.Marshal(),
				conn.SessionID(),
				conn.RemoteAddr().String(),
			)
			if err != nil {
				return nil, err
			}
			if !authResponse.Success {
				return nil, fmt.Errorf("authentication failed")
			}
			return &ssh.Permissions{}, nil
		}
	}

	for index, hostKey := range config.Ssh.HostKeys {
		hostKey = strings.TrimSpace(hostKey)
		if !strings.HasPrefix(hostKey, "-----BEGIN") {
			hostKeyData, err := ioutil.ReadFile(hostKey)
			if err != nil {
				return nil, fmt.Errorf("failed to load host key from %s (%v)", hostKey, err)
			}
			config.Ssh.HostKeys[index] = string(hostKeyData)
		}

		private, err := ssh.ParsePrivateKey([]byte(config.Ssh.HostKeys[index]))
		if err != nil {
			return nil, err
		}

		serverConfig.HostKeys = append(serverConfig.HostKeys, private)
	}

	return server.New(
		config.Listen,
		serverConfig,
		nil,
		NewConnectionHandler(
			config,
			NewDefaultGlobalRequestHandlerFactory(),
			NewDefaultChannelHandlerFactory(
				registry,
				client,
				NewDefaultChannelRequestHandlerFactory(logger),
				logger,
				log.NewLoggerPipelineFactory(logWriter),
			),
		),
		logger,
	)
}
