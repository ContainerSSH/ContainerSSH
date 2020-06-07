package protocol

import "containerssh/config"

type ConfigRequest struct {
	Username  string `json:"username"`
	SessionId string `json:"sessionId"`
}

type ConfigResponse struct {
	Config config.AppConfig `json:"config"`
}
