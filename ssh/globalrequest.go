package ssh

import (
	"context"
	globalRequest "github.com/janoszen/containerssh/ssh/request"
	"github.com/janoszen/containerssh/ssh/server"
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
	Make() server.GlobalRequestHandler
}

type defaultGlobalRequestHandlerFactory struct {
}

func (factory *defaultGlobalRequestHandlerFactory) Make() server.GlobalRequestHandler {
	return &globalRequestHandler{
		globalRequest.NewHandler(),
	}
}

func NewDefaultGlobalRequestHandlerFactory() GlobalRequestHandlerFactory {
	return &defaultGlobalRequestHandlerFactory{}
}
