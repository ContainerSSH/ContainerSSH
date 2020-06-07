package config

import "time"

type AuthConfig struct {
	Url        string        `json:"url" yaml:"url" comment:"URL of the authentication server."`
	CaCert     string        `json:"cacert" yaml:"cacert" comment:"CA certificate file to use for host verification. Note: due to a bug in Go on Windows this has to be explicitly provided."`
	Timeout    time.Duration `json:"timeout" yaml:"timeout" comment:"Timeout in seconds" default:"2s"`
	ClientCert string        `json:"cert" yaml:"cert" comment:"Client certificate file in PEM format"`
	ClientKey  string        `json:"key" yaml:"key" comment:"Client key file in PEM format"`
	Password   bool          `json:"password" yaml:"password" comment:"Perform password authentication" default:"true"`
	PubKey     bool          `json:"pubkey" yaml:"pubkey" comment:"Perform public key authentication" default:"false"`
}
