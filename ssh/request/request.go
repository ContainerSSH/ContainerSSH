package request

import (
	"fmt"
	"github.com/janoszen/containerssh/backend"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type Reply func(success bool, message interface{})

type TypeHandler struct {
	GetRequestObject func() interface{}
	HandleRequest    func(request interface{}, reply Reply, channel ssh.Channel, session backend.Session)
}

type Handler struct {
	typeHandlers map[string]TypeHandler
}

func NewHandler() Handler {
	return Handler{
		typeHandlers: map[string]TypeHandler{},
	}
}

func (handler *Handler) getTypeHandler(requestType string) (*TypeHandler, error) {
	if typeHandler, ok := handler.typeHandlers[requestType]; ok {
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
	channel ssh.Channel,
	session backend.Session,
) {
	typeHandler, err := handler.getTypeHandler(requestType)
	if err != nil {
		log.Println(err)
		reply(false, nil)
	} else if typeHandler == nil {
		reply(false, nil)
	} else {
		typeHandler.HandleRequest(payload, reply, channel, session)
	}
}

func (handler *Handler) AddTypeHandler(requestType string, typeHandler TypeHandler) {
	handler.typeHandlers[requestType] = typeHandler
}

func (handler *Handler) OnRequest(
	requestType string,
	payload []byte,
	reply func(success bool, message []byte),
	channel ssh.Channel,
	session backend.Session,
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

	handler.dispatchRequest(requestType, unmarshalledPayload, replyFunc, channel, session)
}
