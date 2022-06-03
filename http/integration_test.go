package http_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"strings"
	"testing"
	"time"

    "go.containerssh.io/libcontainerssh/config"
    http2 "go.containerssh.io/libcontainerssh/http"
    "go.containerssh.io/libcontainerssh/internal/structutils"
    "go.containerssh.io/libcontainerssh/internal/test"
    "go.containerssh.io/libcontainerssh/log"
    "go.containerssh.io/libcontainerssh/service"
	"github.com/stretchr/testify/assert"
)

type Request struct {
	Message string `json:"Message"`
}

type Response struct {
	Error   bool   `json:"error"`
	Message string `json:"Message"`
}

type handler struct {
}

func (s *handler) OnRequest(request http2.ServerRequest, response http2.ServerResponse) error {
	req := Request{}
	if err := request.Decode(&req); err != nil {
		return err
	}
	if req.Message == "Hi" {
		response.SetBody(&Response{
			Error:   false,
			Message: "Hello world!",
		})
	} else {
		response.SetStatus(400)
		response.SetBody(&Response{
			Error:   true,
			Message: "Be nice and greet me!",
		})
	}
	return nil
}

func TestUnencrypted(t *testing.T) {
	clientConfig, serverConfig := createClientServerConfig(t)

	message := "Hi"

	response, responseStatus, err := runRequest(clientConfig, serverConfig, t, message)
	if err != nil {
		assert.Fail(t, "failed to run request", err)
		return
	}
	assert.Equal(t, 200, responseStatus)
	assert.Equal(t, false, response.Error)
	assert.Equal(t, "Hello world!", response.Message)
}

func createClientServerConfig(t *testing.T) (config.HTTPClientConfiguration, config.HTTPServerConfiguration) {
	clientConfig := config.HTTPClientConfiguration{}
	serverConfig := config.HTTPServerConfiguration{}
	structutils.Defaults(&clientConfig)
	structutils.Defaults(&serverConfig)
	port := test.GetNextPort(t, "config server")
	clientConfig.URL = fmt.Sprintf("http://127.0.0.1:%d/", port)
	serverConfig.Listen = fmt.Sprintf("127.0.0.1:%d", port)
	return clientConfig, serverConfig
}

func TestUnencryptedFailure(t *testing.T) {
	clientConfig, serverConfig := createClientServerConfig(t)

	message := "Hm..."

	response, responseStatus, err := runRequest(clientConfig, serverConfig, t, message)
	if err != nil {
		assert.Fail(t, "failed to run request", err)
		return
	}
	assert.Equal(t, 400, responseStatus)
	assert.Equal(t, true, response.Error)
	assert.Equal(t, "Be nice and greet me!", response.Message)
}

func TestEncrypted(t *testing.T) {
	caPrivKey, caCert, caCertBytes, err := createCA()
	if err != nil {
		assert.Fail(t, "failed to create CA", err)
		return
	}
	serverPrivKey, serverCert, err := createSignedCert(
		[]x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		caPrivKey,
		caCert,
	)
	if err != nil {
		assert.Fail(t, "failed to create server cert", err)
		return
	}

	clientConfig, serverConfig := createClientServerConfig(t)
	//goland:noinspection HttpUrlsUsage
	clientConfig.URL = strings.Replace(clientConfig.URL, "http://", "https://", 1)
	clientConfig.CACert = string(caCertBytes)
	serverConfig.Key = string(serverPrivKey)
	serverConfig.Cert = string(serverCert)

	message := "Hi"

	response, responseStatus, err := runRequest(clientConfig, serverConfig, t, message)
	if err != nil {
		assert.Fail(t, "failed to run request", err)
		return
	}
	assert.Equal(t, 200, responseStatus)
	assert.Equal(t, false, response.Error)
	assert.Equal(t, "Hello world!", response.Message)
}

func TestMutuallyAuthenticated(t *testing.T) {
	caPrivKey, caCert, caCertBytes, err := createCA()
	if err != nil {
		assert.Fail(t, "failed to create CA", err)
		return
	}
	serverPrivKey, serverCert, err := createSignedCert(
		[]x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		caPrivKey,
		caCert,
	)
	if err != nil {
		assert.Fail(t, "failed to create server cert", err)
		return
	}

	clientCaPriv, clientCaCert, clientCaCertBytes, err := createCA()
	if err != nil {
		assert.Fail(t, "failed to create client CA", err)
		return
	}
	clientPrivKey, clientCert, err := createSignedCert(
		[]x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		clientCaPriv,
		clientCaCert,
	)
	if err != nil {
		assert.Fail(t, "failed to create server cert", err)
		return
	}

	clientConfig, serverConfig := createClientServerConfig(t)
	//goland:noinspection HttpUrlsUsage
	clientConfig.URL = strings.Replace(clientConfig.URL, "http://", "https://", 1)
	clientConfig.CACert = string(caCertBytes)
	clientConfig.ClientCert = string(clientCert)
	clientConfig.ClientKey = string(clientPrivKey)
	serverConfig.Key = string(serverPrivKey)
	serverConfig.Cert = string(serverCert)
	serverConfig.ClientCACert = string(clientCaCertBytes)

	message := "Hi"

	response, responseStatus, err := runRequest(clientConfig, serverConfig, t, message)
	if err != nil {
		assert.Fail(t, "failed to run request", err)
		return
	}
	assert.Equal(t, 200, responseStatus)
	assert.Equal(t, false, response.Error)
	assert.Equal(t, "Hello world!", response.Message)
}

func TestMutuallyAuthenticatedFailure(t *testing.T) {
	caPrivKey, caCert, caCertBytes, err := createCA()
	if err != nil {
		assert.Fail(t, "failed to create CA", err)
		return
	}
	serverPrivKey, serverCert, err := createSignedCert(
		[]x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		caPrivKey,
		caCert,
	)
	if err != nil {
		assert.Fail(t, "failed to create server cert", err)
		return
	}

	clientCaPriv, clientCaCert, _, err := createCA()
	if err != nil {
		assert.Fail(t, "failed to create client CA", err)
		return
	}
	clientPrivKey, clientCert, err := createSignedCert(
		[]x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		clientCaPriv,
		clientCaCert,
	)
	if err != nil {
		assert.Fail(t, "failed to create server cert", err)
		return
	}

	clientConfig, serverConfig := createClientServerConfig(t)
	clientConfig.URL = "https://127.0.0.1:8080"
	clientConfig.CACert = string(caCertBytes)
	clientConfig.ClientCert = string(clientCert)
	clientConfig.ClientKey = string(clientPrivKey)
	serverConfig.Key = string(serverPrivKey)
	serverConfig.Cert = string(serverCert)
	//Pass wrong client CA cert to test failure
	serverConfig.ClientCACert = string(caCertBytes)

	message := "Hi"

	if _, _, err = runRequest(clientConfig, serverConfig, t, message); err == nil {
		assert.Fail(t, "Client request with invalid CA verification did not fail.")
		return
	}
}

func createCA() (*rsa.PrivateKey, *x509.Certificate, []byte, error) {
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"ACME, Inc"},
			Country:      []string{"US"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	caPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create private key (%w)", err)
	}
	caCert, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create CA certificate (%w)", err)
	}
	caPEM := new(bytes.Buffer)
	if err := pem.Encode(
		caPEM,
		&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: caCert,
		},
	); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to encode CA cert (%w)", err)
	}
	return caPrivateKey, ca, caPEM.Bytes(), nil
}

func createSignedCert(
	usage []x509.ExtKeyUsage,
	caPrivateKey *rsa.PrivateKey,
	caCertificate *x509.Certificate,
) ([]byte, []byte, error) {
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			Organization: []string{"ACME, Inc"},
			Country:      []string{"US"},
		},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1)},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(0, 0, 1),
		SubjectKeyId: []byte{1},
		ExtKeyUsage:  usage,
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}
	certBytes, err := x509.CreateCertificate(
		rand.Reader,
		cert,
		caCertificate,
		&certPrivKey.PublicKey,
		caPrivateKey,
	)
	if err != nil {
		return nil, nil, err
	}
	certPrivKeyPEM := new(bytes.Buffer)
	if err := pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	}); err != nil {
		return nil, nil, err
	}
	certPEM := new(bytes.Buffer)
	if err := pem.Encode(certPEM,
		&pem.Block{Type: "CERTIFICATE", Bytes: certBytes},
	); err != nil {
		return nil, nil, err
	}
	return certPrivKeyPEM.Bytes(), certPEM.Bytes(), nil
}

func runRequest(
	clientConfig config.HTTPClientConfiguration,
	serverConfig config.HTTPServerConfiguration,
	t *testing.T,
	message string,
) (Response, int, error) {
	response := Response{}
	logger := log.NewTestLogger(t)
	client, err := http2.NewClient(clientConfig, logger)
	if err != nil {
		return response, 0, fmt.Errorf("failed to create client (%w)", err)
	}

	ready := make(chan bool, 1)
	server, err := http2.NewServer(
		"HTTP",
		serverConfig,
		http2.NewServerHandler(&handler{}, logger),
		logger,
		func(_ string) {

		},
	)
	if err != nil {
		return response, 0, fmt.Errorf("failed to create server (%w)", err)
	}
	lifecycle := service.NewLifecycle(server)
	lifecycle.OnRunning(func(s service.Service, l service.Lifecycle) {
		ready <- true
	})

	errorChannel := make(chan error, 2)
	responseStatus := 0
	go func() {
		if err := lifecycle.Run(); err != nil {
			errorChannel <- err
		}
		close(errorChannel)
	}()
	<-ready
	if responseStatus, err = client.Post(
		"",
		&Request{Message: message},
		&response,
	); err != nil {
		lifecycle.Stop(context.Background())
		return response, 0, err
	}
	lifecycle.Stop(context.Background())
	if err, ok := <-errorChannel; ok {
		return response, 0, err
	}
	return response, responseStatus, nil
}
