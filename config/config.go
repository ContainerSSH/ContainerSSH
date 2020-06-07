package config

import "time"

//noinspection GoNameStartsWithPackageName
type ConfigServerConfig struct {
	Timeout    time.Duration `json:"timeout" yaml:"timeout" comment:"HTTP call timeout" default:"2s"`
	Url        string        `json:"url" yaml:"url" comment:"URL of the configuration server. If empty no configuration callout is done."`
	CaCert     string        `json:"cacert" yaml:"cacert" comment:"CA certificate file to use for host verification"`
	ClientCert string        `json:"cert" yaml:"cert" comment:"Client certificate file in PEM format"`
	ClientKey  string        `json:"key" yaml:"key" comment:"Client key file in PEM format"`
}
