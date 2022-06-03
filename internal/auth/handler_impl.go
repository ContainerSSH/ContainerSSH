package auth

import (
	"encoding/base64"
	"fmt"
	goHttp "net/http"
	"strings"

	"go.containerssh.io/libcontainerssh/auth"
	"go.containerssh.io/libcontainerssh/http"
	"go.containerssh.io/libcontainerssh/log"
	"go.containerssh.io/libcontainerssh/message"
)

type handler struct {
	authzHandler    goHttp.Handler
	passwordHandler goHttp.Handler
	pubkeyHandler   goHttp.Handler
}

func (h handler) ServeHTTP(writer goHttp.ResponseWriter, request *goHttp.Request) {
	parts := strings.Split(request.URL.Path, "/")
	switch parts[len(parts)-1] {
	case "authz":
		h.authzHandler.ServeHTTP(writer, request)
	case "password":
		h.passwordHandler.ServeHTTP(writer, request)
	case "pubkey":
		h.pubkeyHandler.ServeHTTP(writer, request)
	default:
		writer.WriteHeader(404)
	}
}

type authzHandler struct {
	backend Handler
	logger  log.Logger
}

func (p *authzHandler) OnRequest(request http.ServerRequest, response http.ServerResponse) error {
	requestObject := auth.AuthorizationRequest{}
	if err := request.Decode(&requestObject); err != nil {
		return err
	}
	success, meta, err := p.backend.OnAuthorization(
		requestObject.ConnectionAuthenticatedMetadata,
	)
	if err != nil {
		p.logger.Debug(message.Wrap(err, message.EAuthRequestDecodeFailed, "failed to execute authorization request"))
		response.SetStatus(500)
		response.SetBody(
			auth.ResponseBody{
				ConnectionAuthenticatedMetadata: requestObject.ConnectionAuthenticatedMetadata,
				Success:                         false,
			},
		)
		return nil
	} else {
		response.SetBody(
			auth.ResponseBody{
				ConnectionAuthenticatedMetadata: meta,
				Success:                         success,
			})
	}
	return nil
}

type passwordHandler struct {
	backend Handler
	logger  log.Logger
}

func (p *passwordHandler) OnRequest(request http.ServerRequest, response http.ServerResponse) error {
	requestObject := auth.PasswordAuthRequest{}
	if err := request.Decode(&requestObject); err != nil {
		return err
	}
	password, err := base64.StdEncoding.DecodeString(requestObject.Password)
	if err != nil {
		return fmt.Errorf("failed to decode password (%w)", err)
	}
	success, meta, err := p.backend.OnPassword(requestObject.ConnectionAuthPendingMetadata, password)
	if err != nil {
		p.logger.Debug(message.Wrap(err, message.EAuthRequestDecodeFailed, "failed to execute password request"))
		response.SetStatus(500)
		response.SetBody(
			auth.ResponseBody{
				ConnectionAuthenticatedMetadata: meta,
				Success:                         false,
			})
		return nil
	} else {
		response.SetBody(
			auth.ResponseBody{
				ConnectionAuthenticatedMetadata: meta,
				Success:                         success,
			})
	}
	return nil
}

type pubKeyHandler struct {
	backend Handler
	logger  log.Logger
}

func (p *pubKeyHandler) OnRequest(request http.ServerRequest, response http.ServerResponse) error {
	requestObject := auth.PublicKeyAuthRequest{}
	if err := request.Decode(&requestObject); err != nil {
		return err
	}
	success, meta, err := p.backend.OnPubKey(
		requestObject.ConnectionAuthPendingMetadata,
		requestObject.PublicKey,
	)
	if err != nil {
		p.logger.Debug(message.Wrap(err, message.EAuthRequestDecodeFailed, "failed to execute public key request"))
		response.SetStatus(500)
		response.SetBody(
			auth.ResponseBody{
				ConnectionAuthenticatedMetadata: meta,
				Success:                         false,
			},
		)
		return nil
	} else {
		response.SetBody(
			auth.ResponseBody{
				ConnectionAuthenticatedMetadata: meta,
				Success:                         success,
			})
	}
	return nil
}
