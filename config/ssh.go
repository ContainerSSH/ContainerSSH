package config

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHConfig is the base configuration structure of the SSH server.
type SSHConfig struct {
	// Listen is the listen address for the SSH server
	Listen string `json:"listen" yaml:"listen" default:"0.0.0.0:2222"`
	// ServerVersion is the version sent to the client.
	//               Must be in the format of "SSH-protoversion-softwareversion SPACE comments".
	//               See https://tools.ietf.org/html/rfc4253#page-4 section 4.2. Protocol Version Exchange
	//               The trailing CR and LF characters should NOT be added to this string.
	ServerVersion SSHServerVersion `json:"serverVersion" yaml:"serverVersion" default:"SSH-2.0-ContainerSSH"`
	// Ciphers are the ciphers offered to the client.
	Ciphers SSHCipherList `json:"ciphers" yaml:"ciphers" default:"[\"chacha20-poly1305@openssh.com\",\"aes256-gcm@openssh.com\",\"aes128-gcm@openssh.com\",\"aes256-ctr\",\"aes192-ctr\",\"aes128-ctr\"]" comment:"SSHCipher suites to use"`
	// KexAlgorithms are the key exchange algorithms offered to the client.
	KexAlgorithms SSHKexList `json:"kex" yaml:"kex" default:"[\"curve25519-sha256@libssh.org\",\"ecdh-sha2-nistp521\",\"ecdh-sha2-nistp384\",\"ecdh-sha2-nistp256\"]" comment:"Key exchange algorithms to use"`
	// MACs are the SSHMAC algorithms offered to the client.
	MACs SSHMACList `json:"macs" yaml:"macs" default:"[\"hmac-sha2-256-etm@openssh.com\",\"hmac-sha2-256\"]" comment:"SSHMAC algorithms to use"`
	// Banner is the banner sent to the client on connecting.
	Banner string `json:"banner" yaml:"banner" comment:"Host banner to show after the username" default:""`
	// HostKeys are the host keys either in PEM format, or filenames to load.
	HostKeys []string `json:"hostkeys" yaml:"hostkeys" comment:"Host keys in PEM format or files to load PEM host keys from."`
	// ClientAliveInterval is the duration between keep alive messages that
	// ContainerSSH will send to each client. If the duration is 0 or unset
	// it disables the feature.
	ClientAliveInterval time.Duration `json:"clientAliveInterval" yaml:"clientAliveInterval" comment:"Inverval to send keepalive packets to the client"`
	// ClientAliveCountMax is the number of keepalive messages that is
	// allowed to be sent without a response being received. If this number
	// is exceeded the connection is considered dead
	ClientAliveCountMax int `json:"clientAliveCountMax" yaml:"clientAliveCountMax" default:"3" comment:"Maximum number of failed keepalives"`
}

// GenerateHostKey generates a random host key and adds it to SSHConfig
func (cfg *SSHConfig) GenerateHostKey() error {
	reader := rand.Reader
	bitSize := 4096
	key, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		return err
	}
	var privateKey = &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}
	var hostKeyBuffer bytes.Buffer
	err = pem.Encode(&hostKeyBuffer, privateKey)
	if err != nil {
		return err
	}

	cfg.HostKeys = append(cfg.HostKeys, hostKeyBuffer.String())
	return nil
}

func (cfg *SSHConfig) LoadHostKeys() ([]ssh.Signer, error) {
	var hostKeys []ssh.Signer
	for index, hostKey := range cfg.HostKeys {
		if strings.TrimSpace(hostKey)[:5] != "-----" {
			// We are deliberalely loading a dynamic file here.
			fh, err := os.Open(hostKey) //nolint:gosec
			if err != nil {
				return nil, fmt.Errorf("failed to load host key %s (%w)", hostKey, err)
			}
			hostKeyData, err := io.ReadAll(fh)
			if err != nil {
				_ = fh.Close()
				return nil, fmt.Errorf("failed to load host key %s (%w)", hostKey, err)
			}
			if err = fh.Close(); err != nil {
				return nil, fmt.Errorf("failed to close host key file %s (%w)", hostKey, err)
			}
			hostKey = string(hostKeyData)
		}
		private, err := ssh.ParsePrivateKey([]byte(hostKey))
		if err != nil {
			return nil, fmt.Errorf("failed to parse host key (%w)", err)
		}
		keyType := private.PublicKey().Type()

		if err := SSHKeyAlgo(keyType).Validate(); err != nil {
			return nil, fmt.Errorf("unsupported host key algorithm %s on host key %d", keyType, index)
		}
		hostKeys = append(hostKeys, private)
	}
	return hostKeys, nil
}

// Validate validates the configuration and returns an error if invalid.
func (cfg SSHConfig) Validate() error {
	if err := cfg.ServerVersion.Validate(); err != nil {
		return wrap(err, "serverVersion")
	}
	if err := cfg.Ciphers.Validate(); err != nil {
		return wrap(err, "ciphers")
	}
	if err := cfg.KexAlgorithms.Validate(); err != nil {
		return wrap(err, "kex")
	}
	if err := cfg.MACs.Validate(); err != nil {
		return wrap(err, "macs")
	}
	if cfg.ClientAliveInterval != 0 && cfg.ClientAliveInterval < 1*time.Second {
		return newError("clientAliveInterval", "clientAliveInterval should be at least 1 second long")
	}
	if cfg.ClientAliveCountMax <= 0 {
		return newError("clientAliveCountMax", "clientAliveCountMax should be at least 1")
	}
	return nil
}

type stringer interface {
	String() string
}

var supportedKexAlgos = []stringer{
	SSHKexCurve25519SHA256, SSHKexECDHSHA2Nistp256, SSHKexECDHSHA2Nistp384, SSHKexECDHSHA2NISTp521,
	SSHKexDHGroup1SHA1, SSHKexDHGroup14SHA1,
}

// SSHKex are the SSH key exchange algorithms
type SSHKex string

// SSHKex are the SSH key exchange algorithms
const (
	SSHKexCurve25519SHA256 SSHKex = "curve25519-sha256@libssh.org"
	SSHKexECDHSHA2NISTp521 SSHKex = "ecdh-sha2-nistp521"
	SSHKexECDHSHA2Nistp384 SSHKex = "ecdh-sha2-nistp384"
	SSHKexECDHSHA2Nistp256 SSHKex = "ecdh-sha2-nistp256"
	SSHKexDHGroup14SHA1    SSHKex = "diffie-hellman-group14-sha1"
	SSHKexDHGroup1SHA1     SSHKex = "diffie-hellman-group1-sha1"
)

// String creates a string representation.
func (k SSHKex) String() string {
	return string(k)
}

// Validate checks if a given SSHKex is valid.
func (k SSHKex) Validate() error {
	if k == "" {
		return fmt.Errorf("empty key exchange algorithm")
	}
	for _, algo := range supportedKexAlgos {
		if algo == k {
			return nil
		}
	}
	return fmt.Errorf("key exchange algorithm not supported: %s", k)
}

// SSHKexList is a list of key exchange algorithms.
type SSHKexList []SSHKex

// Validate validates the key exchange list
func (k SSHKexList) Validate() error {
	if len(k) == 0 {
		return fmt.Errorf("the key exchange list cannot be empty")
	}
	for _, kex := range k {
		if err := kex.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// StringList returns a list of key exchange algorithms as a list.
func (k SSHKexList) StringList() []string {
	result := make([]string, len(k))
	for i, kex := range k {
		result[i] = kex.String()
	}
	return result
}

var supportedHostKeyAlgos = []stringer{
	SSHKeyAlgoSSHRSACertv01, SSHKeyAlgoSSHDSSCertv01, SSHKeyAlgoECDSASHA2NISTp256Certv01,
	SSHKeyAlgoECDSASHA2NISTp384Certv01, SSHKeyAlgoECDSASHA2NISTp521Certv01, SSHKeyAlgoSSHED25519Certv01,
	SSHKeyAlgoSSHRSA, SSHKeyAlgoSSHDSS, SSHKeyAlgoSSHED25519,
}

// SSHKeyAlgo are supported key algorithms.
type SSHKeyAlgo string

// SSHKeyAlgo are supported key algorithms.
const (
	SSHKeyAlgoSSHRSACertv01            SSHKeyAlgo = "ssh-rsa-cert-v01@openssh.com"
	SSHKeyAlgoSSHDSSCertv01            SSHKeyAlgo = "ssh-dss-cert-v01@openssh.com"
	SSHKeyAlgoECDSASHA2NISTp256Certv01 SSHKeyAlgo = "ecdsa-sha2-nistp256-cert-v01@openssh.com"
	SSHKeyAlgoECDSASHA2NISTp384Certv01 SSHKeyAlgo = "ecdsa-sha2-nistp384-cert-v01@openssh.com"
	SSHKeyAlgoECDSASHA2NISTp521Certv01 SSHKeyAlgo = "ecdsa-sha2-nistp521-cert-v01@openssh.com"
	SSHKeyAlgoSSHED25519Certv01        SSHKeyAlgo = "ssh-ed25519-cert-v01@openssh.com"
	SSHKeyAlgoSSHRSA                   SSHKeyAlgo = "ssh-rsa"
	SSHKeyAlgoSSHDSS                   SSHKeyAlgo = "ssh-dss"
	SSHKeyAlgoSSHED25519               SSHKeyAlgo = "ssh-ed25519"
)

// String creates a string representation.
func (h SSHKeyAlgo) String() string {
	return string(h)
}

// Validate checks if a given key algorithm is valid.
func (h SSHKeyAlgo) Validate() error {
	if h == "" {
		return fmt.Errorf("empty host key algorithm")
	}
	for _, algo := range supportedHostKeyAlgos {
		if algo == h {
			return nil
		}
	}
	return fmt.Errorf("unsupported host key algorithm: %s", h)
}

// SSHKeyAlgoList is a list of key algorithms.
type SSHKeyAlgoList []SSHKeyAlgo

// SSHKeyAlgoListFromStringList converts a string list into a list of SSH key algorithms.
func SSHKeyAlgoListFromStringList(hostKeyAlgorithms []string) (SSHKeyAlgoList, error) {
	result := make([]SSHKeyAlgo, len(hostKeyAlgorithms))
	for i, algo := range hostKeyAlgorithms {
		result[i] = SSHKeyAlgo(algo)
	}
	r := SSHKeyAlgoList(result)
	return result, r.Validate()
}

// MustSSHKeyAlgoListFromStringList is identical to SSHKeyAlgoListFromStringList but panics on
// error.
func MustSSHKeyAlgoListFromStringList(hostKeyAlgorithms []string) SSHKeyAlgoList {
	l, err := SSHKeyAlgoListFromStringList(hostKeyAlgorithms)
	if err != nil {
		panic(err)
	}
	return l
}

// Validate validates the list of ciphers to contain only supported items.
func (h SSHKeyAlgoList) Validate() error {
	if len(h) == 0 {
		return fmt.Errorf("host key algorithm list cannot be empty")
	}
	for _, algo := range h {
		if err := algo.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// StringList returns a list of cipher names.
func (h SSHKeyAlgoList) StringList() []string {
	algos := make([]string, len(h))
	for i, v := range h {
		algos[i] = v.String()
	}
	return algos
}

var supportedMACs = []stringer{
	SSHMACHMACSHA2256ETM, SSHMACHMACSHA2256, SSHMACHMACSHA196, SSHMACHMACSHA1,
}

// SSHMAC are the SSH mac algorithms.
type SSHMAC string

// SSHMAC are the SSH mac algorithms.
const (
	SSHMACHMACSHA2256ETM SSHMAC = "hmac-sha2-256-etm@openssh.com"
	SSHMACHMACSHA2256    SSHMAC = "hmac-sha2-256"
	SSHMACHMACSHA1       SSHMAC = "hmac-sha1"
	SSHMACHMACSHA196     SSHMAC = "hmac-sha1-96"
)

// String creates a string representation.
func (m SSHMAC) String() string {
	return string(m)
}

func (m SSHMAC) Validate() error {
	if m == "" {
		return fmt.Errorf("empty SSHMAC")
	}
	for _, algo := range supportedMACs {
		if algo == m {
			return nil
		}
	}
	return fmt.Errorf("SSHMAC not supported: %s", m)
}

// SSHMACList is a list of SSHMAC algorithms
type SSHMACList []SSHMAC

// Validate checks if the SSHMACList is valid.
func (m SSHMACList) Validate() error {
	if len(m) == 0 {
		return fmt.Errorf("empty SSHMAC list")
	}
	for _, mac := range m {
		if err := mac.Validate(); err != nil {
			return fmt.Errorf("invalid SSHMAC (%w)", err)
		}
	}
	return nil
}

// StringList returns a list of SSHMAC names.
func (m SSHMACList) StringList() []string {
	ciphers := make([]string, len(m))
	for i, v := range m {
		ciphers[i] = v.String()
	}
	return ciphers
}

var supportedCiphers = []stringer{
	SSHCipherChaCha20Poly1305, SSHCipherAES256GCM, SSHCipherAES128GCM,
	SSHCipherAES256CTE, SSHCipherAES192CTR, SSHCipherAES128CTR,
	SSHCipherAES128CBC, SSHCipherArcFour256, SSHCipherArcFour128, SSHCipherArcFour, SSHCipherTripleDESCBCID,
}

// SSHCipher is the SSH cipher
type SSHCipher string

// SSHCipher is the SSH cipher
const (
	SSHCipherChaCha20Poly1305 SSHCipher = "chacha20-poly1305@openssh.com"
	SSHCipherAES256GCM        SSHCipher = "aes256-gcm@openssh.com"
	SSHCipherAES128GCM        SSHCipher = "aes128-gcm@openssh.com"
	SSHCipherAES256CTE        SSHCipher = "aes256-ctr"
	SSHCipherAES192CTR        SSHCipher = "aes192-ctr"
	SSHCipherAES128CTR        SSHCipher = "aes128-ctr"
	SSHCipherAES128CBC        SSHCipher = "aes128-cbc"
	SSHCipherArcFour256       SSHCipher = "arcfour256"
	SSHCipherArcFour128       SSHCipher = "arcfour128"
	SSHCipherArcFour          SSHCipher = "arcfour"
	SSHCipherTripleDESCBCID   SSHCipher = "tripledescbcID"
)

// String creates a string representation.
func (c SSHCipher) String() string {
	return string(c)
}

// Validate validates the cipher
func (c SSHCipher) Validate() error {
	if c == "" {
		return fmt.Errorf("empty cipher name")
	}
	for _, supportedCiphers := range supportedCiphers {
		if c == supportedCiphers {
			return nil
		}
	}
	return fmt.Errorf("invalid cipher name: %s", c)
}

// SSHCipherList is a list of supported ciphers
type SSHCipherList []SSHCipher

// Validate validates the list of ciphers to contain only supported items.
func (c SSHCipherList) Validate() error {
	if len(c) == 0 {
		return nil
	}

	for _, cipher := range c {
		if err := cipher.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// StringList returns a list of cipher names.
func (c SSHCipherList) StringList() []string {
	ciphers := make([]string, len(c))
	for i, v := range c {
		ciphers[i] = v.String()
	}
	return ciphers
}

var serverVersionRegexp = regexp.MustCompile(`^SSH-2.0-[!-,.-~]+( [a-zA-Z0-9- _.]+)?$`)
var serverVersionMaxLen = 253

// SSHServerVersion is a string that is issued to the client when connecting.
type SSHServerVersion string

// Validate checks if the server version conforms to RFC 4253 section 4.2.
// See https://tools.ietf.org/html/rfc4253#page-4
func (s SSHServerVersion) Validate() error {
	if len(s) > serverVersionMaxLen || !serverVersionRegexp.MatchString(string(s)) {
		return fmt.Errorf("invalid server version string (%s), see https://tools.ietf.org/html/rfc4253#page-4 section 4.2. for details", s)
	}
	return nil
}

// String returns a string from the SSHServerVersion.
func (s SSHServerVersion) String() string {
	return string(s)
}
