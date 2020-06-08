package client

import "github.com/janoszen/containerssh/protocol"

type ConfigClient interface {
	GetConfig(request protocol.ConfigRequest) (*protocol.ConfigResponse, error)
}
