package request

import (
	"fmt"
	"github.com/containerssh/containerssh/audit"
	"github.com/containerssh/containerssh/audit/protocol"
	"github.com/containerssh/containerssh/backend"
	"github.com/containerssh/containerssh/log"
	"golang.org/x/crypto/ssh"
)

type Reply func(success bool, message interface{})

type TypeHandler interface {
	GetRequestObject() interface{}
	HandleRequest(request interface{}, reply Reply, channel ssh.Channel, session backend.Session, auditChannel *audit.Channel)
}

type Handler struct {
	channelHandlers map[string]TypeHandler
	logger          log.Logger
}

func NewHandler(logger log.Logger) Handler {
	return Handler{
		channelHandlers: map[string]TypeHandler{},
		logger:          logger,
	}
}

func (handler *Handler) getTypeHandler(requestType string) (TypeHandler, error) {
	if typeHandler, ok := handler.channelHandlers[requestType]; ok {
		return typeHandler, nil
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
	auditChannel *audit.Channel,
) {
	typeHandler, err := handler.getTypeHandler(requestType)
	if err != nil {
		handler.logger.InfoE(err)
		auditChannel.Message(protocol.MessageType_UnknownChannelRequestType, &protocol.MessageUnknownChannelRequestType{RequestType: requestType})
		reply(false, nil)
	} else if typeHandler == nil {
		auditChannel.Message(protocol.MessageType_UnknownChannelRequestType, &protocol.MessageUnknownChannelRequestType{RequestType: requestType})
		reply(false, nil)
	} else {
		typeHandler.HandleRequest(payload, reply, channel, session, auditChannel)
	}
}

func (handler *Handler) AddTypeHandler(requestType string, typeHandler TypeHandler) {
	handler.channelHandlers[requestType] = typeHandler
}

func (handler *Handler) OnChannelRequest(requestType string, payload []byte, reply func(success bool, message []byte), channel ssh.Channel, session backend.Session, auditChannel *audit.Channel) {
	unmarshalledPayload, err := handler.getPayloadObjectForRequestType(requestType)
	if err != nil {
		auditChannel.Message(protocol.MessageType_UnknownChannelRequestType, protocol.MessageUnknownChannelRequestType{RequestType: requestType})
		handler.logger.InfoE(err)
		reply(false, nil)
	}

	if payload != nil && len(payload) > 0 {
		err = ssh.Unmarshal(payload, unmarshalledPayload)
		if err != nil {
			auditChannel.Message(protocol.MessageType_FailedToDecodeChannelRequest, protocol.MessageUnknownChannelRequestType{RequestType: requestType})
			handler.logger.InfoE(err)
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

	handler.dispatchRequest(requestType, unmarshalledPayload, replyFunc, channel, session, auditChannel)
}
