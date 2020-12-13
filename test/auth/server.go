package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	testHttp "github.com/containerssh/containerssh/test/http"
	"github.com/containerssh/containerssh/test/protocol"
)

type MemoryAuthServer struct {
	passwords map[string]string
	keys      map[string][]string
	http      *testHttp.Server
	mutex     *sync.Mutex
}

func NewMemoryAuthServer() *MemoryAuthServer {
	httpServer := testHttp.New(8080)

	server := &MemoryAuthServer{
		passwords: make(map[string]string, 0),
		keys:      make(map[string][]string),
		http:      httpServer,
		mutex:     &sync.Mutex{},
	}

	httpServer.GetMux().HandleFunc("/pubkey", server.authPublicKey)
	httpServer.GetMux().HandleFunc("/password", server.authPassword)

	return server
}

func (server *MemoryAuthServer) SetPassword(username string, password string) {
	server.passwords[username] = password
}

func (server *MemoryAuthServer) AddPublicKey(username string, publicKeyBase64 string) {
	server.mutex.Lock()
	if _, ok := server.keys[username]; !ok {
		server.keys[username] = []string{}
	}
	server.keys[username] = append(server.keys[username], publicKeyBase64)
	server.mutex.Unlock()
}

func (server *MemoryAuthServer) Start() error {
	return server.http.Start()
}

func (server *MemoryAuthServer) Stop() error {
	if server.http != nil {
		return server.http.Stop()
	} else {
		return fmt.Errorf("http already stopped")
	}
}

func (server *MemoryAuthServer) authPassword(w http.ResponseWriter, req *http.Request) {
	var authRequest protocol.PasswordAuthRequest
	err := json.NewDecoder(req.Body).Decode(&authRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	authResponse := protocol.AuthResponse{
		Success: false,
	}

	if base64.StdEncoding.EncodeToString([]byte(server.passwords[authRequest.Username])) == authRequest.Password {
		authResponse.Success = true
	}

	_ = json.NewEncoder(w).Encode(authResponse)
}

func (server *MemoryAuthServer) authPublicKey(w http.ResponseWriter, req *http.Request) {
	var authRequest protocol.PublicKeyAuthRequest
	err := json.NewDecoder(req.Body).Decode(&authRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	authResponse := protocol.AuthResponse{
		Success: false,
	}

	if val, ok := server.keys[authRequest.Username]; ok {
		for _, key := range val {
			if key == authRequest.PublicKey {
				authResponse.Success = true
			}
		}
	}

	_ = json.NewEncoder(w).Encode(authResponse)
}
