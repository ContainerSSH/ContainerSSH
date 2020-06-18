package ssh

type ServerConfiguration struct {
	Listen        string   `json:"listen" yaml:"listen" comment:"IP and port to listen on" default:"0.0.0.0:22"`
	HostKeys      []string `json:"hostkeys" yaml:"hostkeys" comment:"Host key files to use. Files must be in PEM format."`
	Ciphers       []string `json:"ciphers" yaml:"ciphers" default:"[\"chacha20-poly1305@openssh.com\",\"aes256-gcm@openssh.com\",\"aes128-gcm@openssh.com\",\"aes256-ctr\",\"aes192-ctr\",\"aes128-ctr\"]" comment:"Cipher suites to use"`
	KexAlgorithms []string `json:"kex" yaml:"kex" default:"[\"curve25519-sha256@libssh.org\",\"ecdh-sha2-nistp521\",\"ecdh-sha2-nistp384\",\"ecdh-sha2-nistp256\"]" comment:"Key exchange algorithms to use"`
	Macs          []string `json:"macs" yaml:"macs" default:"[\"hmac-sha2-256-etm@openssh.com\",\"hmac-sha2-256\",\"hmac-sha1\",\"hmac-sha1-96\"]" comment:"MAC algorithms to use"`
}
