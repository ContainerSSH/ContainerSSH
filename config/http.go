package config

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"strings"
	"time"
)

// TLSVersion is the version of the TLS protocol to use.
type TLSVersion string

const (
	TLSVersion12 TLSVersion = "1.2"
	TLSVersion13 TLSVersion = "1.3"
)

// Validate validates the TLS version
func (t TLSVersion) Validate() error {
	switch t {
	case TLSVersion13:
		fallthrough
	case TLSVersion12:
		return nil
	default:
		return fmt.Errorf("unsupported TLS version: %s", t)
	}
}

// GetTLSVersion returns the go-representation of the TLS version.
func (t TLSVersion) GetTLSVersion() uint16 {
	switch t {
	case TLSVersion13:
		return tls.VersionTLS12
	case TLSVersion12:
		return tls.VersionTLS12
	default:
		panic(fmt.Errorf("invalid TLS version: %s", t))
	}
}

// ECDHCurveList is a list of supported ECDHCurve
type ECDHCurveList []ECDHCurve

// Validate provides validation for a list of cipher suites.
func (c ECDHCurveList) Validate() error {
	for _, curve := range c {
		if err := curve.Validate(); err != nil {
			return err
		}
	}
	if len(c) == 0 {
		return fmt.Errorf("no ECDH curves provided")
	}
	return nil
}

// ECDHCurve is an elliptic curve algorithm.
type ECDHCurve string

// Elliptic curve algorithms.
const (
	ECDHCurveX25519       ECDHCurve = "x25519"
	ECDHCurveX25519Alt    ECDHCurve = "X25519"
	ECDHCurveSecP256r1    ECDHCurve = "secp256r1"
	ECDHCurveSecP256r1Alt ECDHCurve = "prime256v1"
	ECDHCurveSecP384r1    ECDHCurve = "secp384r1"
	ECDHCurveSecP521r1    ECDHCurve = "secp521r1"
)

var curveToID = map[ECDHCurve]tls.CurveID{
	ECDHCurveX25519:       tls.X25519,
	ECDHCurveX25519Alt:    tls.X25519,
	ECDHCurveSecP256r1:    tls.CurveP256,
	ECDHCurveSecP256r1Alt: tls.CurveP256,
	ECDHCurveSecP384r1:    tls.CurveP384,
	ECDHCurveSecP521r1:    tls.CurveP521,
}

// Validate validates the TLS curve for a valid value.
func (c ECDHCurve) Validate() error {
	switch c {
	case ECDHCurveX25519:
	case ECDHCurveX25519Alt:
	case ECDHCurveSecP256r1:
	case ECDHCurveSecP256r1Alt:
	case ECDHCurveSecP384r1:
	case ECDHCurveSecP521r1:
	default:
		return fmt.Errorf("invalid ECDH curve: %s", c)
	}
	return nil
}

func (c ECDHCurve) getCurveID() tls.CurveID {
	if curveID, ok := curveToID[c]; ok {
		return curveID
	}
	panic(fmt.Errorf("invalid ECDH curve: %s", c))
}

// UnmarshalJSON provides JSON unmarshalling from both a list and a string with ECDH curves.
func (c *ECDHCurveList) UnmarshalJSON(data []byte) error {
	var decoded []ECDHCurve
	if err := json.Unmarshal(data, &decoded); err != nil {
		var decodedString string
		if err := json.Unmarshal(data, &decodedString); err != nil {
			return fmt.Errorf("failed to unmarshal ECDH curve list, neither a list nor a string is provided")
		}
		for _, entry := range strings.Split(decodedString, ":") {
			*c = append(*c, ECDHCurve(entry))
		}
		return nil
	}
	for _, entry := range decoded {
		*c = append(*c, entry)
	}
	return nil
}

// UnmarshalYAML provides YAML unmarshalling from both a list and a string with ECDH curves.
func (c *ECDHCurveList) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var decoded []ECDHCurve
	if err := unmarshal(&decoded); err != nil {
		var decodedString string
		if err := unmarshal(&decodedString); err != nil {
			return fmt.Errorf("failed to unmarshal ECDH curve list, neither a list nor a string is provided")
		}
		for _, entry := range strings.Split(decodedString, ":") {
			*c = append(*c, ECDHCurve(entry))
		}
		return nil
	}
	for _, entry := range decoded {
		*c = append(*c, entry)
	}
	return nil
}

// GetList returns a list of go-internal TLS CurveIDs.
func (c ECDHCurveList) GetList() []tls.CurveID {
	var result []tls.CurveID
	for _, s := range c {
		result = append(result, s.getCurveID())
	}
	return result
}

// CipherSuiteList is a list of cipher suites. This type is provided for easier unmarshaling from a list or string.
type CipherSuiteList []CipherSuite

// UnmarshalJSON provides JSON unmarshalling from both a list and a cipher suite string.
func (c *CipherSuiteList) UnmarshalJSON(data []byte) error {
	var decoded []CipherSuite
	if err := json.Unmarshal(data, &decoded); err != nil {
		var decodedString string
		if err := json.Unmarshal(data, &decodedString); err != nil {
			return fmt.Errorf("failed to unmarshal cipher suite list, neither a list nor a string is provided")
		}
		for _, entry := range strings.Split(decodedString, ":") {
			*c = append(*c, CipherSuite(entry))
		}
		return nil
	}
	for _, entry := range decoded {
		*c = append(*c, entry)
	}
	return nil
}

// UnmarshalYAML provides YAML unmarshalling from both a list and a cipher suite string.
func (c *CipherSuiteList) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var decoded []CipherSuite
	if err := unmarshal(&decoded); err != nil {
		var decodedString string
		if err := unmarshal(&decodedString); err != nil {
			return fmt.Errorf("failed to unmarshal cipher suite list, neither a list nor a string is provided")
		}
		for _, entry := range strings.Split(decodedString, ":") {
			*c = append(*c, CipherSuite(entry))
		}
		return nil
	}
	for _, entry := range decoded {
		*c = append(*c, entry)
	}
	return nil
}

// Validate provides validation for a list of cipher suites.
func (c CipherSuiteList) Validate() error {
	for _, suite := range c {
		if err := suite.Validate(); err != nil {
			return err
		}
	}
	if len(c) == 0 {
		return fmt.Errorf("no cipher suites provided")
	}
	return nil
}

// GetList returns a go-specific list of cipher IDs.
func (c CipherSuiteList) GetList() []uint16 {
	var result []uint16
	for _, s := range c {
		result = append(result, s.getCipher())
	}
	return result
}

// CipherSuite is the cipher suite used for TLS connections.
type CipherSuite string

const (
	IANA_TLS_AES_128_GCM_SHA256                     CipherSuite = "TLS_AES_128_GCM_SHA256"
	IANA_TLS_AES_256_GCM_SHA384                     CipherSuite = "TLS_AES_256_GCM_SHA384"
	IANA_TLS_CHACHA20_POLY1305_SHA256               CipherSuite = "TLS_CHACHA20_POLY1305_SHA256"
	IANA_TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256    CipherSuite = "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256"
	OpenSSL_TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256 CipherSuite = "ECDHE-ECDSA-AES128-GCM-SHA256"
	GnuTLS_TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256  CipherSuite = "TLS_ECDHE_ECDSA_AES_128_GCM_SHA256"
	IANA_TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256      CipherSuite = "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
	OpenSSL_TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256   CipherSuite = "ECDHE-RSA-AES128-GCM-SHA256"
	GnuTLS_TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256    CipherSuite = "TLS_ECDHE_RSA_AES_128_GCM_SHA256"
	IANA_TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384    CipherSuite = "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384"
	OpenSSL_TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384 CipherSuite = "ECDHE-ECDSA-AES256-GCM-SHA384"
	GnuTLS_TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384  CipherSuite = "TLS_ECDHE_ECDSA_AES_256_GCM_SHA384"
	IANA_TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384      CipherSuite = "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"
	OpenSSL_TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384   CipherSuite = "ECDHE-RSA-AES256-GCM-SHA384"
	GnuTLS_TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384    CipherSuite = "TLS_ECDHE_RSA_AES_256_GCM_SHA384"
	IANA_TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305     CipherSuite = "TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305"
	OpenSSL_TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305  CipherSuite = "ECDHE-ECDSA-CHACHA20-POLY1305"
	GnuTLS_TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305   CipherSuite = "TLS_ECDHE_ECDSA_CHACHA20_POLY1305"
	IANA_TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305       CipherSuite = "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305"
	OpenSSL_TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305    CipherSuite = "ECDHE-RSA-CHACHA20-POLY1305"
	GnuTLS_TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305     CipherSuite = "TLS_ECDHE_RSA_CHACHA20_POLY1305"
)

var stringToCipherSuite = map[CipherSuite]uint16{
	IANA_TLS_AES_128_GCM_SHA256:                  tls.TLS_AES_128_GCM_SHA256,
	IANA_TLS_AES_256_GCM_SHA384:                  tls.TLS_AES_256_GCM_SHA384,
	IANA_TLS_CHACHA20_POLY1305_SHA256:            tls.TLS_CHACHA20_POLY1305_SHA256,
	IANA_TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256: tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	IANA_TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256:   tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	IANA_TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384: tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	IANA_TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384:   tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	IANA_TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305:  tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
	IANA_TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305:    tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,

	OpenSSL_TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256: tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	OpenSSL_TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256:   tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	OpenSSL_TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384: tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	OpenSSL_TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384:   tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	OpenSSL_TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305:  tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
	OpenSSL_TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305:    tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,

	GnuTLS_TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256: tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	GnuTLS_TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256:   tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	GnuTLS_TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384: tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	GnuTLS_TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384:   tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	GnuTLS_TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305:  tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
	GnuTLS_TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305:    tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
}

// Validate validates if the cipher suite is supported.
func (c CipherSuite) Validate() error {
	if _, ok := stringToCipherSuite[c]; !ok {
		return fmt.Errorf("unsupported cipher suite: %s", c)
	}
	return nil
}

func (c CipherSuite) getCipher() uint16 {
	if suite, ok := stringToCipherSuite[c]; ok {
		return suite
	}
	panic(fmt.Errorf("unsupported cipher suite: %s", c))
}

// HTTPClientConfiguration is the configuration structure for HTTP clients
//
//We are adding the JSON and YAML tags to conform to the Operator SDK requirements to tag all fields.
//goland:noinspection GoVetStructTag
type HTTPClientConfiguration struct {
	// URL is the base URL for requests.
	URL string `json:"url" yaml:"url" comment:"Base URL of the server to connect."`

	// AllowRedirects sets if the client should honor HTTP redirects. Defaults to false.
	AllowRedirects bool `json:"allowRedirects" yaml:"allowRedirects" comment:""`

	// Timeout is the time the client should wait for a response.
	Timeout time.Duration `json:"timeout" yaml:"timeout" comment:"HTTP call timeout." default:"2s"`

	// CACert is either the CA certificate to expect on the server in PEM format
	//         or the name of a file containing the PEM.
	CACert string `json:"cacert" yaml:"cacert" comment:"CA certificate in PEM format to use for host verification."`

	// ClientCert is a PEM containing an x509 certificate to present to the server or a file name containing the PEM.
	ClientCert string `json:"cert" yaml:"cert" comment:"Client certificate file in PEM format."`

	// ClientKey is a PEM containing a private key to use to connect the server or a file name containing the PEM.
	ClientKey string `json:"key" yaml:"key" comment:"Client key file in PEM format."`

	// TLSVersion is the minimum TLS version to use.
	TLSVersion TLSVersion `json:"tlsVersion" yaml:"tlsVersion" default:"1.3"`

	// ECDHCurves is the list of curve algorithms to support.
	ECDHCurves ECDHCurveList `json:"curves" yaml:"curves" default:"[\"x25519\",\"secp256r1\",\"secp384r1\",\"secp521r1\"]"`

	// CipherSuites is a list of supported cipher suites.
	CipherSuites CipherSuiteList `json:"cipher" yaml:"cipher" default:"[\"TLS_AES_128_GCM_SHA256\",\"TLS_AES_256_GCM_SHA384\",\"TLS_CHACHA20_POLY1305_SHA256\",\"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256\",\"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256\"]"`

	// RequestEncoding is the means by which the request body is encoded. It defaults to JSON encoding.
	RequestEncoding RequestEncoding `json:"-" yaml:"-"`
}

// HTTPClientCerts is a structure that holds the client certificates after successfully calling ValidateWithCerts
type HTTPClientCerts struct {
	// CACertPool contains the certificate pool used for verifying the server identity.
	CACertPool *x509.CertPool
	// Cert is the client certificate to authenticate to the server with, if any.
	Cert *tls.Certificate
}

// Validate validates the HTTP client configuration.
func (c *HTTPClientConfiguration) Validate() error {
	_, err := c.ValidateWithCerts()
	return err
}

// ValidateWithCerts validates the client configuration and returns the loaded certificates if any.
func (c *HTTPClientConfiguration) ValidateWithCerts() (*HTTPClientCerts, error) {
	result := &HTTPClientCerts{}

	_, err := url.ParseRequestURI(c.URL)
	if err != nil {
		return nil, newError("url", "invalid URL: %s", c.URL)
	}
	if c.Timeout < 100*time.Millisecond {
		return nil, newError("timeout", "timeout value %s is too low, must be at least 100ms", c.Timeout.String())
	}

	if err := c.validateCACert(result); err != nil {
		return nil, wrap(err, "cacert")
	}

	if err := c.RequestEncoding.Validate(); err != nil {
		return nil, err
	}

	if strings.HasPrefix(c.URL, "https://") {
		if err := c.TLSVersion.Validate(); err != nil {
			return nil, wrap(err, "tlsVersion")
		}
		if err := c.ECDHCurves.Validate(); err != nil {
			return nil, wrap(err, "curves")
		}
		if err := c.CipherSuites.Validate(); err != nil {
			return nil, wrap(err, "cipher")
		}
	}

	err = c.validateClientCert(result)
	return result, err
}

func (c *HTTPClientConfiguration) validateClientCert(certs *HTTPClientCerts) error {
	if c.ClientCert != "" && c.ClientKey == "" {
		return newError("clientCert", "client certificate provided without client key")
	} else if c.ClientCert == "" && c.ClientKey != "" {
		return newError("clientKey", "client key provided without client certificate")
	}

	if c.ClientCert != "" && c.ClientKey != "" {
		clientCert, err := loadPEM(c.ClientCert)
		if err != nil {
			return wrapWithMessage(err, "clientCert", "failed to load client certificate")
		}
		clientKey, err := loadPEM(c.ClientKey)
		if err != nil {
			return wrapWithMessage(err, "clientKey", "failed to load client certificate")
		}
		cert, err := tls.X509KeyPair(clientCert, clientKey)
		if err != nil {
			return wrapWithMessage(err, "clientCert", "failed to load certificate or key")
		}
		certs.Cert = &cert
	}
	return nil
}

func (c *HTTPClientConfiguration) validateCACert(certs *HTTPClientCerts) (err error) {
	if strings.TrimSpace(c.CACert) != "" {
		caCert, err := loadPEM(c.CACert)
		if err != nil {
			return fmt.Errorf("failed to load CA certificate (%w)", err)
		}

		certs.CACertPool = x509.NewCertPool()
		if !certs.CACertPool.AppendCertsFromPEM(caCert) {
			return fmt.Errorf("invalid CA certificate provided")
		}
	} else if strings.HasPrefix(c.URL, "https://") {
		certs.CACertPool, err = x509.SystemCertPool()
		if err != nil {
			return fmt.Errorf(
				"system certificate pool unusable and no explicit CA certificate was given (%w)",
				err,
			)
		}
	}
	return nil
}

// HTTPServerConfiguration is a structure to configure the simple HTTP server by.
//goland:noinspection GoVetStructTag
type HTTPServerConfiguration struct {
	// Listen contains the IP and port to listen on.
	Listen string `json:"listen" yaml:"listen" default:"0.0.0.0:8080"`
	// Key contains either a file name to a private key, or the private key itself in PEM format to use as a server key.
	Key string `json:"key" yaml:"key"`
	// Cert contains either a file to a certificate, or the certificate itself in PEM format to use as a server
	// certificate.
	Cert string `json:"cert" yaml:"cert"`
	// ClientCACert contains either a file or a certificate in PEM format to verify the connecting clients by.
	ClientCACert string `json:"clientcacert" yaml:"clientcacert"`

	// TLSVersion is the minimum TLS version to use.
	TLSVersion TLSVersion `json:"tlsVersion" yaml:"tlsVersion" default:"1.3"`

	// ECDHCurves is the list of curve algorithms to support.
	ECDHCurves ECDHCurveList `json:"curves" yaml:"curves" default:"[\"x25519\",\"secp256r1\",\"secp384r1\",\"secp521r1\"]"`

	// CipherSuites is a list of supported cipher suites.
	CipherSuites CipherSuiteList `json:"cipher" yaml:"cipher" default:"[\"TLS_AES_128_GCM_SHA256\",\"TLS_AES_256_GCM_SHA384\",\"TLS_CHACHA20_POLY1305_SHA256\",\"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256\",\"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256\"]"`
}

func (config *HTTPServerConfiguration) Validate() error {
	_, err := config.ValidateWithCerts()
	return err
}

// HTTPServerCerts is a structure returned from ValidateWithCerts, containing the loaded certificates after validation.
type HTTPServerCerts struct {
	// Cert is the server certificate.
	Cert *tls.Certificate
	// ClientCAPool contains the CA certificate pool to verify client certificates against, if any.
	ClientCAPool *x509.CertPool
}

// ValidateWithCerts validates the server configuration and returns the loaded certificates
func (config *HTTPServerConfiguration) ValidateWithCerts() (*HTTPServerCerts, error) {
	if config.Listen == "" {
		return nil, fmt.Errorf("no listen address provided")
	}
	if _, _, err := net.SplitHostPort(config.Listen); err != nil {
		return nil, fmt.Errorf("invalid listen address provided (%w)", err)
	}
	if config.Cert != "" && config.Key == "" {
		return nil, fmt.Errorf("certificate provided without a key")
	}
	if config.Cert == "" && config.Key != "" {
		return nil, fmt.Errorf("key provided without certificate")
	}

	result := &HTTPServerCerts{}

	if config.Cert != "" && config.Key != "" {
		pemCert, err := loadPEM(config.Cert)
		if err != nil {
			return nil, fmt.Errorf("failed to load certificate (%w)", err)
		}
		pemKey, err := loadPEM(config.Key)
		if err != nil {
			return nil, fmt.Errorf("failed to load key (%w)", err)
		}
		cert, err := tls.X509KeyPair(pemCert, pemKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load key/certificate (%w)", err)
		}
		result.Cert = &cert

		if err := config.TLSVersion.Validate(); err != nil {
			return nil, fmt.Errorf("invalid TLS version (%w)", err)
		}
		if err := config.ECDHCurves.Validate(); err != nil {
			return nil, fmt.Errorf("invalid curve algorithms (%w)", err)
		}
		if err := config.CipherSuites.Validate(); err != nil {
			return nil, fmt.Errorf("invalid cipher suites (%w)", err)
		}
	}

	if config.ClientCACert != "" {
		clientCaCert, err := loadPEM(config.ClientCACert)
		if err != nil {
			return nil, fmt.Errorf("failed to load client CA certificate (%w)", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(clientCaCert) {
			return nil, fmt.Errorf("failed to load client CA certificate")
		}
		result.ClientCAPool = caCertPool
	}

	return result, nil
}

// RequestEncoding is the method by which the response is encoded.
type RequestEncoding string

// RequestEncodingDefault is the default encoding and encodes the body to JSON.
const RequestEncodingDefault = ""

// RequestEncodingJSON encodes the body to JSON.
const RequestEncodingJSON = "JSON"

// RequestEncodingWWWURLEncoded encodes the body via www-urlencoded.
const RequestEncodingWWWURLEncoded = "WWW-URLENCODED"

// Validate validates the RequestEncoding
func (r RequestEncoding) Validate() error {
	switch r {
	case RequestEncodingDefault:
		return nil
	case RequestEncodingJSON:
		return nil
	case RequestEncodingWWWURLEncoded:
		return nil
	default:
		return fmt.Errorf("invalid request encoding: %s", r)
	}
}

func loadPEM(spec string) ([]byte, error) {
	if !strings.HasPrefix(strings.TrimSpace(spec), "-----") {
		// We are deliberately reading a file here.
		return ioutil.ReadFile(spec) //nolint:gosec
	}
	return []byte(spec), nil
}
