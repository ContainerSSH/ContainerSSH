package request

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type Reply func(success bool, message interface{})

type TypeHandler struct {
	GetRequestObject func() interface{}
	HandleRequest    func(request interface{}, reply Reply)
}

type Handler struct {
	GlobalHandlers map[string]TypeHandler
}

func NewHandler() *Handler {
	return &Handler{
		GlobalHandlers: map[string]TypeHandler{},
	}
}

func (handler *Handler) getTypeHandler(requestType string) (*TypeHandler, error) {
	if typeHandler, ok := handler.GlobalHandlers[requestType]; ok {
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
		log.Println(err)
		reply(false, nil)
	} else if typeHandler == nil {
		reply(false, nil)
	} else {
		typeHandler.HandleRequest(payload, reply)
	}
}

func (handler *Handler) AddTypeHandler(requestType string, typeHandler TypeHandler) {
	handler.GlobalHandlers[requestType] = typeHandler
}

func (handler *Handler) OnGlobalRequest(
	requestType string,
	payload []byte,
	reply func(success bool, message []byte),
) {
	unmarshalledPayload, err := handler.getPayloadObjectForRequestType(requestType)
	if err != nil {
		log.Println(err)
		reply(false, nil)
	}

	if payload != nil && len(payload) > 0 {
		err = ssh.Unmarshal(payload, unmarshalledPayload)
		if err != nil {
			log.Println(err)
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
