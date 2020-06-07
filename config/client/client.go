package client

import "containerssh/protocol"

type ConfigClient interface {
	GetConfig(request protocol.ConfigRequest) (*protocol.ConfigResponse, error)
}
