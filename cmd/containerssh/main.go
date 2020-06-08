package main

import (
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"github.com/janoszen/containerssh/auth"
	"github.com/janoszen/containerssh/backend"
	"github.com/janoszen/containerssh/backend/dockerrun"
	"github.com/janoszen/containerssh/config"
	configurationClient "github.com/janoszen/containerssh/config/client"
	"github.com/janoszen/containerssh/config/loader"
	"github.com/janoszen/containerssh/config/util"
	"github.com/janoszen/containerssh/protocol"
	containerssh "github.com/janoszen/containerssh/ssh"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"net"
	"os"
)

func InitBackendRegistry() *backend.Registry {
	registry := backend.NewRegistry()
	dockerrun.Init(registry)
	return registry
}

func getSshConfig(appConfig *config.AppConfig, authClient auth.Client) (*ssh.ServerConfig, error) {
	sshConfig := &ssh.ServerConfig{}

	if appConfig.Auth.Password {
		sshConfig.PasswordCallback = func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
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
				return nil, errors.New("authentication failed")
			}
			return &ssh.Permissions{}, nil
		}
	}

	if appConfig.Auth.PubKey {
		sshConfig.PublicKeyCallback = func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
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
				return nil, errors.New("authentication failed")
			}
			return &ssh.Permissions{}, nil
		}
	}

	if len(appConfig.Ssh.HostKeys) == 0 {
		return nil, fmt.Errorf("no host keys provided")
	}

	for _, hostKeyFile := range appConfig.Ssh.HostKeys {
		hostKeyData, err := ioutil.ReadFile(hostKeyFile)
		if err != nil {
			return nil, err
		}
		signer, err := ssh.ParsePrivateKey(hostKeyData)
		if err != nil {
			return nil, err
		}
		sshConfig.AddHostKey(signer)
	}

	sshConfig.KeyExchanges = appConfig.Ssh.KexAlgorithms
	sshConfig.MACs = appConfig.Ssh.Macs
	sshConfig.Ciphers = appConfig.Ssh.Ciphers

	if sshConfig.PublicKeyCallback == nil && sshConfig.PasswordCallback == nil {
		return nil, fmt.Errorf("neither public key nor password authentication is configured")
	}
	if len(sshConfig.KeyExchanges) == 0 {
		return nil, fmt.Errorf("no key exchanges configured")
	}
	if len(sshConfig.MACs) == 0 {
		return nil, fmt.Errorf("no key MACs configured")
	}
	if len(sshConfig.Ciphers) == 0 {
		return nil, fmt.Errorf("no key ciphers configured")
	}
	return sshConfig, nil
}

func main() {

	backendRegistry := InitBackendRegistry()
	appConfig, err := util.GetDefaultConfig()
	if err != nil {
		log.Fatal(err)
	}

	configFile := ""
	dumpConfig := false
	flag.StringVar(&configFile, "config", "", "Configuration file to load (has to end in .yaml, .yml, or .json)")
	flag.BoolVar(&dumpConfig, "dump-config", false, "Dump configuration and exit")
	flag.Parse()

	if configFile != "" {
		fileAppConfig, err := loader.LoadFile(configFile)
		if err != nil {
			log.Fatal(err)
		}
		appConfig, err = util.Merge(fileAppConfig, appConfig)
		if err != nil {
			log.Fatal(err)
		}
	}

	if dumpConfig {
		err := loader.Write(appConfig, os.Stdout)
		if err != nil {
			log.Fatal(err)
		} else {
			os.Exit(0)
		}
	}

	authClient, err := auth.NewHttpAuthClient(appConfig.Auth)
	if err != nil {
		log.Fatal(err)
	}

	configClient, err := configurationClient.NewHttpConfigClient(appConfig.ConfigServer)
	if err != nil {
		log.Fatal(err)
	}

	sshConfig, err := getSshConfig(appConfig, authClient)
	if err != nil {
		log.Fatal(err)
	}

	listener, err := net.Listen("tcp", appConfig.Listen)
	if err != nil {
		log.Fatalf("Failed to listen on %s (%s)", appConfig.Listen, err)
	}

	log.Printf("Listening on %s", appConfig.Listen)
	for {
		tcpConn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept incoming connection (%s)", err)
			continue
		}
		sshConn, chans, reqs, err := ssh.NewServerConn(tcpConn, sshConfig)
		if err != nil {
			log.Printf("Failed to handshake (%s)", err)
			continue
		}

		log.Printf("New SSH connection from %s (%s)", sshConn.RemoteAddr(), sshConn.ClientVersion())
		go ssh.DiscardRequests(reqs)
		go handleChannels(sshConn, chans, appConfig, backendRegistry, configClient)
	}
}

func handleChannels(conn *ssh.ServerConn, chans <-chan ssh.NewChannel, appConfig *config.AppConfig, backendRegistry *backend.Registry, configClient configurationClient.ConfigClient) {
	for newChannel := range chans {
		go handleChannel(conn, newChannel, appConfig, backendRegistry, configClient)
	}
}

func handleChannel(conn *ssh.ServerConn, newChannel ssh.NewChannel, appConfig *config.AppConfig, backendRegistry *backend.Registry, configClient configurationClient.ConfigClient) {
	if t := newChannel.ChannelType(); t != "session" {
		_ = newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
		return
	}

	actualConfig := appConfig

	if configClient != nil {
		configResponse, err := configClient.GetConfig(protocol.ConfigRequest{
			Username:  conn.User(),
			SessionId: base64.StdEncoding.EncodeToString(conn.SessionID()),
		})
		if err != nil {
			log.Print(err)
			_ = newChannel.Reject(ssh.ResourceShortage, fmt.Sprintf("internal error while calling the config server: %s", err))
			return
		}
		actualConfig, err = util.Merge(&configResponse.Config, appConfig)
		if err != nil {
			log.Print(err)
			_ = newChannel.Reject(ssh.ResourceShortage, fmt.Sprintf("failed to merge config: %s", err))
			return
		}
	}

	containerBackend, err := backendRegistry.GetBackend(actualConfig.Backend)
	if err != nil {
		log.Print(err)
		_ = newChannel.Reject(ssh.ResourceShortage, fmt.Sprintf("no such backend: %s", err))
		return
	}

	backendSession, err := containerBackend.CreateSession(
		string(conn.SessionID()),
		conn.User(),
		actualConfig,
	)
	if err != nil {
		_ = newChannel.Reject(ssh.ConnectionFailed, fmt.Sprintf("internal error while creating backend: %s", err))
		return
	}

	connection, requests, err := newChannel.Accept()
	if err != nil {
		log.Printf("Could not accept channel (%s)", err)
		backendSession.Close()
		return
	}

	closeConnections := func() {
		backendSession.Close()
		err := connection.Close()
		if err != nil {
			log.Fatal(err)
		}
	}

	requestHandlers := containerssh.InitRequestHandlers()

	go func() {
		for req := range requests {
			reply := func(success bool, message []byte) {
				if req.WantReply {
					err := req.Reply(success, message)
					if err != nil {
						closeConnections()
					}
				}
			}
			requestHandlers.OnRequest(req.Type, req.Payload, reply, connection, backendSession)
		}
	}()
}
