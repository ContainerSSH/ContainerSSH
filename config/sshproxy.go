package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHProxyConfig is the configuration for the SSH proxy module.
type SSHProxyConfig struct {
	// Server is the IP address or hostname of the backing server.
	Server string `json:"server" yaml:"server"`
	// Port is the TCP port to connect to.
	Port uint16 `json:"port" yaml:"port" default:"22"`
	// UsernamePassThrough means that the username should be taken from the connecting client.
	UsernamePassThrough bool `json:"usernamePassThrough" yaml:"usernamePassThrough"`
	// Username is the username to pass to the backing SSH server for authentication.
	Username string `json:"username" yaml:"username"`
	// Password is the password to offer to the backing SSH server for authentication.
	Password string `json:"password" yaml:"password"`
	// PrivateKey is the private key to use for authenticating with the backing server.
	PrivateKey string `json:"privateKey" yaml:"privateKey"`
	// AllowedHostKeyFingerprints lists which fingerprints we accept
	AllowedHostKeyFingerprints SSHProxyAllowedHostKeyFingerprints `json:"allowedHostKeyFingerprints" yaml:"allowedHostKeyFingerprints"`
	// Ciphers are the ciphers supported for the backend connection.
	Ciphers SSHCipherList `json:"ciphers" yaml:"ciphers" default:"[\"chacha20-poly1305@openssh.com\",\"aes256-gcm@openssh.com\",\"aes128-gcm@openssh.com\",\"aes256-ctr\",\"aes192-ctr\",\"aes128-ctr\"]" comment:"Cipher suites to use"`
	// KexAlgorithms are the key exchange algorithms for the backend connection.
	KexAlgorithms SSHKexList `json:"kex" yaml:"kex" default:"[\"curve25519-sha256@libssh.org\",\"ecdh-sha2-nistp521\",\"ecdh-sha2-nistp384\",\"ecdh-sha2-nistp256\"]" comment:"Key exchange algorithms to use"`
	// MACs are the MAC algorithms for the backend connection.
	MACs SSHMACList `json:"macs" yaml:"macs" default:"[\"hmac-sha2-256-etm@openssh.com\",\"hmac-sha2-256\"]" comment:"MAC algorithms to use"`
	// HostKeyAlgorithms is a list of algorithms for host keys. The server can offer multiple host keys and this list
	// are the ones we want to accept. The fingerprints for the accepted algorithms should be added to
	// AllowedHostKeyFingerprints.
	HostKeyAlgorithms SSHKeyAlgoList `json:"hostKeyAlgos" yaml:"hostKeyAlgos" default:"[\"ssh-rsa-cert-v01@openssh.com\",\"ssh-dss-cert-v01@openssh.com\",\"ecdsa-sha2-nistp256-cert-v01@openssh.com\",\"ecdsa-sha2-nistp384-cert-v01@openssh.com\",\"ecdsa-sha2-nistp521-cert-v01@openssh.com\",\"ssh-ed25519-cert-v01@openssh.com\",\"ssh-rsa\",\"ssh-dss\",\"ssh-ed25519\"]"`
	// Timeout is the time ContainerSSH is willing to wait for the backing connection to be established.
	Timeout time.Duration `json:"timeout" yaml:"timeout" default:"60s"`
	// ClientVersion is the version sent to the server.
	//               Must be in the format of "SSH-protoversion-softwareversion SPACE comments".
	//               See https://tools.ietf.org/html/rfc4253#page-4 section 4.2. Protocol Version Exchange
	//               The trailing CR and LF characters should NOT be added to this string.
	ClientVersion SSHProxyClientVersion `json:"clientVersion" yaml:"clientVersion" default:"SSH-2.0-ContainerSSH"`
}

// Validate checks the configuration for the backing SSH server.
func (c SSHProxyConfig) Validate() error {
	if c.Server == "" {
		return newError("server", "server cannot be empty")
	}
	if c.Port == 0 || c.Port > 65535 {
		return newError("port", "invalid port number: %d", c.Port)
	}
	if c.Username == "" && !c.UsernamePassThrough {
		return newError("username", "username cannot be empty when usernamePassThrough is not set")
	}
	if len(c.AllowedHostKeyFingerprints) == 0 {
		return newError("allowedHostKeyFingerprints", "allowedHostKeyFingerprints cannot be empty")
	}
	if err := c.Ciphers.Validate(); err != nil {
		return wrap(err, "ciphers")
	}
	if err := c.KexAlgorithms.Validate(); err != nil {
		return wrap(err, "kex")
	}
	if err := c.MACs.Validate(); err != nil {
		return wrap(err, "macs")
	}
	if err := c.HostKeyAlgorithms.Validate(); err != nil {
		return wrap(err, "hostKeyAlgos")
	}
	if err := c.ClientVersion.Validate(); err != nil {
		return wrap(err, "clientVersion")
	}
	return nil
}

func (c SSHProxyConfig) LoadPrivateKey() (ssh.Signer, error) {
	if c.PrivateKey == "" {
		return nil, nil
	}
	privateKey := c.PrivateKey
	if strings.TrimSpace(privateKey)[:5] != "-----" {
		// Loading file here, so no gosec problems.
		fh, err := os.Open(privateKey) //nolint:gosec
		if err != nil {
			return nil, fmt.Errorf("failed load private key %s (%w)", privateKey, err)
		}
		privateKeyData, err := ioutil.ReadAll(fh)
		if err != nil {
			_ = fh.Close()
			return nil, fmt.Errorf("failed to load private key %s (%w)", privateKey, err)
		}
		if err = fh.Close(); err != nil {
			return nil, fmt.Errorf("failed to close host key file %s (%w)", privateKey, err)
		}
		privateKey = string(privateKeyData)
	}
	private, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key (%w)", err)
	}
	keyType := private.PublicKey().Type()

	if err := SSHKeyAlgo(keyType).Validate(); err != nil {
		return nil, fmt.Errorf("unsupported host key algorithm %s", keyType)
	}
	return private, nil
}

var clientVersionRegexp = regexp.MustCompile(`^SSH-2.0-[a-zA-Z0-9]+(| [a-zA-Z0-9- _.]+)$`)

// SSHProxyClientVersion is a string that is issued to the client when connecting.
type SSHProxyClientVersion string

// Validate checks if the client version conforms to RFC 4253 section 4.2.
// See https://tools.ietf.org/html/rfc4253#page-4
func (s SSHProxyClientVersion) Validate() error {
	if !clientVersionRegexp.MatchString(string(s)) {
		return fmt.Errorf("invalid client version string (%s), see https://tools.ietf.org/html/rfc4253#page-4 section 4.2. for details", s)
	}
	return nil
}

// String returns a string from the SSHProxyClientVersion.
func (s SSHProxyClientVersion) String() string {
	return string(s)
}

var fingerprintValidator = regexp.MustCompile("^SSH256:[a-zA-Z0-9/+]+$")

// SSHProxyAllowedHostKeyFingerprints is a list of fingerprints that ContainerSSH is allowed to connect to.
type SSHProxyAllowedHostKeyFingerprints []string

// Validate validates the correct format of the host key fingerprints.
func (a SSHProxyAllowedHostKeyFingerprints) Validate() error {
	if len(a) == 0 {
		return fmt.Errorf("no host keys provided")
	}
	for _, fp := range a {
		if !fingerprintValidator.Match([]byte(fp)) {
			return fmt.Errorf("invalid fingerprint: %s (must start with SHA256:)", fp)
		}
	}
	return nil
}
