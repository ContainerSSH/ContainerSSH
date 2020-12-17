package config

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/containerssh/configuration"

	testHttp "github.com/containerssh/containerssh/test/http"
	"github.com/containerssh/containerssh/test/protocol"
)

type MemoryConfigServer struct {
	defaultConfig configuration.AppConfig
	userConfig    map[string]configuration.AppConfig
	server        *testHttp.Server
	mutex         *sync.Mutex
}

func NewMemoryConfigServer() *MemoryConfigServer {
	httpServer := testHttp.New(8081)

	server := &MemoryConfigServer{
		defaultConfig: configuration.AppConfig{},
		userConfig:    make(map[string]configuration.AppConfig),
		server:        httpServer,
		mutex:         &sync.Mutex{},
	}

	httpServer.GetMux().HandleFunc("/config", server.config)

	return server
}

func (server *MemoryConfigServer) config(w http.ResponseWriter, req *http.Request) {
	var configRequest protocol.ConfigRequest
	err := json.NewDecoder(req.Body).Decode(&configRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := protocol.ConfigResponse{
		Config: server.defaultConfig,
	}

	if val, ok := server.userConfig[configRequest.Username]; ok {
		response.Config = val
	}

	_ = json.NewEncoder(w).Encode(response)
}

func (server *MemoryConfigServer) Start() error {
	return server.server.Start()
}

func (server *MemoryConfigServer) Stop() error {
	return server.server.Stop()
}

func (server *MemoryConfigServer) SetUserConfig(
	username string,
	configuration *configuration.AppConfig,
) {
	server.userConfig[username] = *configuration
}
