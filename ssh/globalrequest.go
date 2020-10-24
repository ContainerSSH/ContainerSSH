package ssh

import (
	"context"
	"github.com/containerssh/containerssh/audit"
	globalRequest "github.com/containerssh/containerssh/ssh/request"
	"github.com/containerssh/containerssh/ssh/server"
	"golang.org/x/crypto/ssh"
)

type globalRequestHandler struct {
	typeHandler *globalRequest.Handler
}

func (handler *globalRequestHandler) OnGlobalRequest(
	ctx context.Context,
	_ *ssh.ServerConn,
	requestType string,
	payload []byte,
) server.RequestResponse {
	responseChannel := make(chan server.RequestResponse)
	reply := func(success bool, message []byte) {
		responseChannel <- server.RequestResponse{
			Success: success,
			Payload: message,
		}
	}
	go handler.typeHandler.OnGlobalRequest(
		requestType,
		payload,
		reply,
	)
	select {
	case response := <-responseChannel:
		return response
	case <-ctx.Done():
		return server.RequestResponse{
			Success: false,
			Payload: []byte("server is shutting down"),
		}
	}
}

type GlobalRequestHandlerFactory interface {
	Make(auditConnection *audit.Connection) server.GlobalRequestHandler
}

type defaultGlobalRequestHandlerFactory struct {
}

func (factory *defaultGlobalRequestHandlerFactory) Make(auditConnection *audit.Connection) server.GlobalRequestHandler {
	return &globalRequestHandler{
		globalRequest.NewHandler(auditConnection),
	}
}

func NewDefaultGlobalRequestHandlerFactory() GlobalRequestHandlerFactory {
	return &defaultGlobalRequestHandlerFactory{}
}
