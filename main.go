package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/docker/docker/api/types"
	containerType "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

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

type passwordAuthRequest struct {
	User          string `json:"user"`
	RemoteAddress string `json:"remoteAddress"`
	SessionId     string `json:"sessionIdBase64"`
	Password      string `json:"passwordBase64"`
}

type publicKeyAuthRequest struct {
	User          string `json:"user"`
	RemoteAddress string `json:"remoteAddress"`
	SessionId     string `json:"sessionIdBase64"`
	// serialized key data in SSH wire format
	PublicKey string `json:"publicKeyBase64"`
}

type authResponse struct {
	Success     bool            `json:"success"`
	Permissions ssh.Permissions `json:"permissions"`
}

func authServerRequest(authServer string, requestObject interface{}, response interface{}) error {
	httpClient := http.Client{
		Timeout: time.Second * 2, // Maximum of 2 secs
	}
	buffer := &bytes.Buffer{}
	json.NewEncoder(buffer).Encode(requestObject)
	req, err := http.NewRequest(http.MethodGet, authServer, buffer)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		return err
	}
	return nil
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
	var dockerHost string

	flag.StringVar(&ciphers, "ciphers", "chacha20-poly1305@openssh.com,aes256-gcm@openssh.com,aes128-gcm@openssh.com,aes256-ctr,aes192-ctr,aes128-ctr", "Supported ciphers")
	flag.StringVar(&kexAlgos, "kex-algorithms", "curve25519-sha256@libssh.org,ecdh-sha2-nistp521,ecdh-sha2-nistp384,ecdh-sha2-nistp256", "Key exchange algorithms")
	flag.StringVar(&macs, "macs", "hmac-sha2-512-etm@openssh.com,hmac-sha2-256-etm@openssh.com,umac-128-etm@openssh.com,hmac-sha2-512,hmac-sha2-256,umac-128@openssh.com", "MAC methods")
	flag.StringVar(&hostKeyRSA, "hostkey-rsa", "", "RSA host key file")
	flag.StringVar(&hostKeyECDSA, "hostkey-ecdsa", "", "ECDSA host key file")
	flag.StringVar(&hostKeyED25519, "hostkey-ed25519", "", "ED25519 host key file")
	flag.StringVar(&listen, "listen", "0.0.0.0:22", "IP and port to listen on")
	flag.StringVar(&authServer, "auth-server", "http://localhost:8080", "HTTP server that will answer authentication requests.")
	flag.BoolVar(&authPubkey, "auth-pubkey", false, "Authenticate using public keys")
	flag.BoolVar(&authPassword, "auth-password", false, "Authenticate using passwords")
	flag.StringVar(&dockerHost, "docker-host", "tcp://127.0.0.1:2375", "Docker TCP socket")

	flag.Parse()

	config := &ssh.ServerConfig{}

	if !authPassword && !authPubkey {
		log.Fatal("Neither Password nor SSH key authentication is selected. Cowardly refusing to let everyone in.")
	}

	if authPassword {
		config.PasswordCallback = func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			authRequest := passwordAuthRequest{
				User:          conn.User(),
				RemoteAddress: conn.RemoteAddr().String(),
				SessionId:     base64.StdEncoding.EncodeToString(conn.SessionID()),
				Password:      base64.StdEncoding.EncodeToString(password),
			}
			authResponse := &authResponse{}
			err := authServerRequest(authServer+"/password", authRequest, authResponse)
			if err != nil {
				return nil, err
			}
			if !authResponse.Success {
				return nil, errors.New("authentication failed")
			}
			return &authResponse.Permissions, nil
		}
	}

	if authPubkey {
		config.PublicKeyCallback = func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			authRequest := publicKeyAuthRequest{
				User:          conn.User(),
				RemoteAddress: conn.RemoteAddr().String(),
				SessionId:     base64.StdEncoding.EncodeToString(conn.SessionID()),
				PublicKey:     base64.StdEncoding.EncodeToString(key.Marshal()),
			}
			authResponse := &authResponse{}
			err := authServerRequest(authServer+"/pubkey", authRequest, authResponse)
			if err != nil {
				return nil, err
			}
			if !authResponse.Success {
				return nil, errors.New("authentication failed")
			}
			return &authResponse.Permissions, nil
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
		go handleChannels(chans)
	}
}

func handleChannels(chans <-chan ssh.NewChannel) {
	for newChannel := range chans {
		go handleChannel(newChannel)
	}
}

func handleChannel(newChannel ssh.NewChannel) {
	if t := newChannel.ChannelType(); t != "session" {
		newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
		return
	}

	connection, requests, err := newChannel.Accept()
	if err != nil {
		log.Printf("Could not accept channel (%s)", err)
		return
	}

	ctx := context.Background()
	cli, err := client.NewClient(dockerHost, "", nil, make(map[string]string))
	if err != nil {
		log.Printf("Could not open connection to Docker server (%s)", err)
		connection.Close()
		return
	}

	containerConfig := &containerType.Config{}
	containerConfig.Image = "busybox"
	containerConfig.Hostname = "test"
	containerConfig.Domainname = "example.com"
	containerConfig.AttachStdin = true
	containerConfig.AttachStderr = true
	containerConfig.AttachStdout = true
	containerConfig.OpenStdin = true
	containerConfig.StdinOnce = true
	containerConfig.Tty = true
	containerConfig.NetworkDisabled = false
	containerConfig.Cmd = []string{"/bin/sh"}
	hostConfig := &containerType.HostConfig{}
	hostConfig.AutoRemove = true
	networkingConfig := &network.NetworkingConfig{}
	body, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, networkingConfig, "")
	if err != nil {
		log.Printf("Failed to create container (%s)", err)
		connection.Close()
		cli.Close()
		return
	}
	containerId := body.ID
	close := func() {
		//todo handle errors
		removeOptions := types.ContainerRemoveOptions{Force: true}
		cli.ContainerRemove(ctx, containerId, removeOptions)
		connection.Close()
		//todo handle errors
		cli.Close()
		log.Printf("Session closed")
	}
	attachOptions := types.ContainerAttachOptions{
		Logs: true,
		Stdin: true,
		Stderr: true,
		Stdout: true,
		Stream: true,
	}
	attachResult, err := cli.ContainerAttach(ctx, containerId, attachOptions)
	if err != nil {
		log.Printf("Failed to attach container (%s)", err)
		close()
		return
	}
	extendedClose := func() {
		attachResult.Close()
		_ = attachResult.CloseWrite()
		close()
	}


	startOptions := types.ContainerStartOptions{}
	err = cli.ContainerStart(ctx, containerId, startOptions)
	if err != nil{
		log.Printf("Failed to start container (%s)", err)
		extendedClose()
		return
	}

	var once sync.Once
	go func() {
		io.Copy(connection, attachResult.Reader)
		once.Do(extendedClose)
	}()
	go func() {
		io.Copy(attachResult.Conn, connection)
		once.Do(extendedClose)
	}()

	go func() {
		for req := range requests {
			switch req.Type {
			case "shell":
				if len(req.Payload) == 0 {
					req.Reply(true, nil)
				}
			case "pty-req":
				termLen := req.Payload[3]

				w, h := parseDims(req.Payload[termLen+4:])
				resizeOptions := types.ResizeOptions{}
				resizeOptions.Width = uint(w)
				resizeOptions.Height = uint(h)
				cli.ContainerResize(ctx, containerId, resizeOptions)
				req.Reply(true, nil)
			case "window-change":
				w, h := parseDims(req.Payload)
				resizeOptions := types.ResizeOptions{}
				resizeOptions.Width = uint(w)
				resizeOptions.Height = uint(h)
				cli.ContainerResize(ctx, containerId, resizeOptions)
			}
		}
	}()
}

func parseDims(b []byte) (uint32, uint32) {
	w := binary.BigEndian.Uint32(b)
	h := binary.BigEndian.Uint32(b[4:])
	return w, h
}
