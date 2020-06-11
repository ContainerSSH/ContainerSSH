package http

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"runtime"
	"strings"
	"time"
)

func NewHttpClient(
	Timeout time.Duration,
	CaCert string,
	ClientCert string,
	ClientKey string,
	Url string,
) (*http.Client, error) {
	tlsConfig := &tls.Config{}
	if CaCert != "" {
		caCert, err := ioutil.ReadFile(CaCert)
		if err != nil {
			return nil, err
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
	} else if runtime.GOOS == "windows" && strings.HasPrefix(Url, "https://") {
		//Remove if https://github.com/golang/go/issues/16736 gets fixed
		return nil, fmt.Errorf("due to a bug (#16736) in Golang on Windows CA certificates have to be explicitly provided for https:// authentication server URLs")
	}

	if ClientCert != "" && ClientKey != "" {
		cert, err := tls.LoadX509KeyPair(ClientCert, ClientKey)
		if err != nil {
			log.Fatal(err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}
	transport := &http.Transport{TLSClientConfig: tlsConfig}

	return &http.Client{
		Transport: transport,
		Timeout:   Timeout,
	}, nil
}
