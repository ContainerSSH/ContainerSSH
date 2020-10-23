package client

import "github.com/containerssh/containerssh/protocol"

type ConfigClient interface {
	GetConfig(request protocol.ConfigRequest) (*protocol.ConfigResponse, error)
}
