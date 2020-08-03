package http

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
)

type Server struct {
	ctx    context.Context
	cancel context.CancelFunc
	srv    *http.Server
	mutex  *sync.Mutex
	mux    *http.ServeMux
	port   int
}

func New(port int) *Server {
	return &Server{
		port: port,
		mutex: &sync.Mutex{},
		mux:   http.NewServeMux(),
	}
}

func (server *Server) GetMux() * http.ServeMux {
	return server.mux
}

func (server *Server) monitorAndRun() {
	go server.run()

	server.mutex.Lock()
	if server.ctx != nil {
		ctx := server.ctx
		server.mutex.Unlock()
		<-ctx.Done()
	} else {
		server.mutex.Unlock()
	}
	server.mutex.Lock()
	if server.srv != nil {
		_ = server.srv.Shutdown(context.TODO())
		server.srv = nil
		server.ctx = nil
	}
	server.mutex.Unlock()
}

func (server *Server) run() {
	server.mutex.Lock()
	srv := &http.Server{Addr: "127.0.0.1:" + strconv.Itoa(server.port), Handler: server.mux}
	server.srv = srv
	server.mutex.Unlock()
	err := srv.ListenAndServe()
	if err != nil {
		server.mutex.Lock()
		server.cancel()
		server.ctx = nil
		server.cancel = nil
		server.mutex.Unlock()
		return
	}

	server.mutex.Lock()
	server.ctx = nil
	server.cancel = nil
	server.mutex.Unlock()
}

func (server *Server) Start() error {
	server.mutex.Lock()
	if server.ctx == nil {
		server.ctx, server.cancel = context.WithCancel(context.Background())
		server.mutex.Unlock()
		go server.monitorAndRun()
		return nil
	} else {
		server.mutex.Unlock()
		return fmt.Errorf("server is already running")
	}
}

func (server *Server) Stop() error {
	server.mutex.Lock()
	if server.cancel != nil {
		server.cancel()
		server.mutex.Unlock()
		return nil
	} else {
		server.mutex.Unlock()
		return fmt.Errorf("server is not running")
	}
}