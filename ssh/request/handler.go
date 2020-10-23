package request

import (
	"fmt"

	"github.com/containerssh/containerssh/log"

	"golang.org/x/crypto/ssh"
)

type Reply func(success bool, message interface{})

type TypeHandler struct {
	GetRequestObject func() interface{}
	HandleRequest    func(request interface{}, reply Reply)
}

type Handler struct {
	globalHandlers map[string]TypeHandler
	logger         log.Logger
}

func NewHandler() *Handler {
	return &Handler{
		globalHandlers: map[string]TypeHandler{},
	}
}

func (handler *Handler) getTypeHandler(requestType string) (*TypeHandler, error) {
	if typeHandler, ok := handler.globalHandlers[requestType]; ok {
		return &typeHandler, nil
	}
	return nil, fmt.Errorf("unsupported request type: %s", requestType)
}

func (handler *Handler) getPayloadObjectForRequestType(requestType string) (interface{}, error) {
	typeHandler, err := handler.getTypeHandler(requestType)
	if err != nil {
		return nil, err
	}
	return typeHandler.GetRequestObject(), nil
}

func (handler *Handler) dispatchRequest(
	requestType string,
	payload interface{},
	reply Reply,
) {
	typeHandler, err := handler.getTypeHandler(requestType)
	if err != nil {
		handler.logger.DebugE(err)
		reply(false, nil)
	} else if typeHandler == nil {
		reply(false, nil)
	} else {
		typeHandler.HandleRequest(payload, reply)
	}
}

func (handler *Handler) AddTypeHandler(requestType string, typeHandler TypeHandler) {
	handler.globalHandlers[requestType] = typeHandler
}

func (handler *Handler) OnGlobalRequest(
	requestType string,
	payload []byte,
	reply func(success bool, message []byte),
) {
	unmarshalledPayload, err := handler.getPayloadObjectForRequestType(requestType)
	if err != nil {
		handler.logger.DebugE(err)
		reply(false, nil)
	}

	if payload != nil && len(payload) > 0 {
		err = ssh.Unmarshal(payload, unmarshalledPayload)
		if err != nil {
			handler.logger.DebugE(err)
			reply(false, nil)
		}
	}

	replyFunc := func(success bool, message interface{}) {
		if message != nil {
			reply(success, ssh.Marshal(message))
		} else {
			reply(success, nil)
		}
	}

	handler.dispatchRequest(requestType, unmarshalledPayload, replyFunc)
}
