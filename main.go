package main

import (
	"containerssh/auth"
	"containerssh/backend"
	"containerssh/backend/dockerrun"
	containerssh "containerssh/ssh"
	"errors"
	"flag"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"net"
	"strings"
)

func InitBackendRegistry() *backend.Registry {
	registry := backend.NewRegistry()
	dockerrun.Init(registry)
	return registry
}

func addHostKey(hostKeyFile string, config *ssh.ServerConfig) {
	if hostKeyFile == "" {
		return
	}
	privateBytes, err := ioutil.ReadFile(hostKeyFile)
	if err != nil {
		log.Fatalf("Failed to load private key (%s)", hostKeyFile)
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		log.Fatal("Failed to parse private key")
	}

	config.AddHostKey(private)
}

func main() {
	var ciphers string
	var kexAlgos string
	var macs string
	var hostKeyRSA string
	var hostKeyECDSA string
	var hostKeyED25519 string
	var listen string
	var authServer string
	var authPubkey bool
	var authPassword bool
	var selectedBackendKey string

	registry := InitBackendRegistry()

	flag.StringVar(&ciphers, "ciphers", "chacha20-poly1305@openssh.com,aes256-gcm@openssh.com,aes128-gcm@openssh.com,aes256-ctr,aes192-ctr,aes128-ctr", "Supported ciphers")
	flag.StringVar(&kexAlgos, "kex-algorithms", "curve25519-sha256@libssh.org,ecdh-sha2-nistp521,ecdh-sha2-nistp384,ecdh-sha2-nistp256", "Key exchange algorithms")
	flag.StringVar(&macs, "macs", "hmac-sha2-256-etm@openssh.com,hmac-sha2-256,hmac-sha1,hmac-sha1-96", "MAC methods")
	flag.StringVar(&hostKeyRSA, "hostkey-rsa", "", "RSA host key file")
	flag.StringVar(&hostKeyECDSA, "hostkey-ecdsa", "", "ECDSA host key file")
	flag.StringVar(&hostKeyED25519, "hostkey-ed25519", "", "ED25519 host key file")
	flag.StringVar(&listen, "listen", "0.0.0.0:22", "IP and port to listen on")
	flag.StringVar(&authServer, "auth-server", "http://localhost:8080", "HTTP server that will answer authentication requests.")
	flag.BoolVar(&authPubkey, "auth-pubkey", false, "Authenticate using public keys")
	flag.BoolVar(&authPassword, "auth-password", false, "Authenticate using passwords")
	flag.StringVar(&selectedBackendKey, "backend", registry.GetBackends()[0], fmt.Sprintf("Backend to use (%s)", strings.Join(registry.GetBackends(), ",")))

	flag.Parse()

	selectedBackend, err := registry.GetBackend(selectedBackendKey)
	if err != nil {
		log.Fatal(err)
	}

	authClient := auth.NewHttpClient(authServer)
	config := &ssh.ServerConfig{}

	if !authPassword && !authPubkey {
		log.Fatal("Neither Password nor SSH key authentication is selected. Cowardly refusing to let everyone in.")
	}

	if authPassword {
		config.PasswordCallback = func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
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

	if authPubkey {
		config.PublicKeyCallback = func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
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

	if hostKeyRSA == "" && hostKeyECDSA == "" && hostKeyED25519 == "" {
		log.Fatal("No host key supplied. Please supply at least one host key.")
	}

	config.KeyExchanges = strings.Split(kexAlgos, ",")
	config.MACs = strings.Split(macs, ",")
	config.Ciphers = strings.Split(ciphers, ",")

	if len(config.KeyExchanges) == 0 {
		log.Fatal("No key exchanges configured.")
	}

	if len(config.MACs) == 0 {
		log.Fatal("No key MACs configured.")
	}

	if len(config.Ciphers) == 0 {
		log.Fatal("No key ciphers configured.")
	}

	addHostKey(hostKeyRSA, config)
	addHostKey(hostKeyECDSA, config)
	addHostKey(hostKeyED25519, config)

	listener, err := net.Listen("tcp", listen)
	if err != nil {
		log.Fatalf("Failed to listen on %s (%s)", listen, err)
	}

	log.Printf("Listening on %s", listen)
	for {
		tcpConn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept incoming connection (%s)", err)
			continue
		}
		sshConn, chans, reqs, err := ssh.NewServerConn(tcpConn, config)
		if err != nil {
			log.Printf("Failed to handshake (%s)", err)
			continue
		}

		log.Printf("New SSH connection from %s (%s)", sshConn.RemoteAddr(), sshConn.ClientVersion())
		go ssh.DiscardRequests(reqs)
		go handleChannels(sshConn, chans, selectedBackend)
	}
}

func handleChannels(conn *ssh.ServerConn, chans <-chan ssh.NewChannel, backend *backend.Backend) {
	for newChannel := range chans {
		go handleChannel(conn, newChannel, backend)
	}
}

func handleChannel(conn *ssh.ServerConn, newChannel ssh.NewChannel, backend *backend.Backend) {
	if t := newChannel.ChannelType(); t != "session" {
		_ = newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
		return
	}

	backendSession, err := backend.CreateSession(
		string(conn.SessionID()),
		conn.User(),
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
