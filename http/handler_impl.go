package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	goHttp "net/http"
	"sort"
	"strconv"
	"strings"

    log2 "go.containerssh.io/libcontainerssh/log"
    "go.containerssh.io/libcontainerssh/message"
)

type serverResponse struct {
	statusCode uint16
	body       interface{}
}

func (s *serverResponse) Error() string {
	return fmt.Sprintf("%v", s.body)
}

func (s *serverResponse) SetStatus(statusCode uint16) {
	s.statusCode = statusCode
}

func (s *serverResponse) SetBody(body interface{}) {
	s.body = body
}

type handler struct {
	requestHandler            RequestHandler
	logger                    log2.Logger
	defaultResponseMarshaller responseMarshaller
	defaultResponseType       string
	responseMarshallers       []responseMarshaller
}

var internalErrorResponse = serverResponse{
	500,
	map[string]string{"error": "Internal Server Error"},
}

var badRequestResponse = serverResponse{
	400,
	map[string]string{"error": "Bad Request"},
}

func (h *handler) ServeHTTP(goWriter goHttp.ResponseWriter, goRequest *goHttp.Request) {
	response := serverResponse{
		statusCode: 200,
		body:       nil,
	}
	if err := h.requestHandler.OnRequest(
		&internalRequest{
			request: goRequest,
			writer:  goWriter,
		},
		&response,
	); err != nil {
		if errors.Is(err, &badRequestResponse) {
			response = badRequestResponse
		} else {
			response = internalErrorResponse
		}
	}
	marshaller, responseType, statusCode := h.findMarshaller(goWriter, goRequest)
	if statusCode != 200 {
		goWriter.WriteHeader(statusCode)
	}
	bytes, err := marshaller.Marshal(response.body)
	if err != nil {
		h.logger.Error(message.Wrap(err, message.MHTTPServerEncodeFailed, "failed to marshal response %v", response))
		response = internalErrorResponse
		bytes, err = json.Marshal(internalErrorResponse.body)
		if err != nil {
			// This should never happen
			panic(fmt.Errorf("bug: failed to marshal internal server error JSON response (%w)", err))
		}
	}
	goWriter.WriteHeader(int(response.statusCode))
	goWriter.Header().Add("Content-Type", responseType)
	if _, err := goWriter.Write(bytes); err != nil {
		h.logger.Debug(message.Wrap(err, message.MHTTPServerResponseWriteFailed, "Failed to write HTTP response"))
	}
}

func (h *handler) findMarshaller(_ goHttp.ResponseWriter, request *goHttp.Request) (responseMarshaller, string, int) {
	acceptHeader := request.Header.Get("Accept")
	if acceptHeader == "" {
		return h.defaultResponseMarshaller, h.defaultResponseType, 200
	}

	accepted := strings.Split(acceptHeader, ",")
	acceptMap := make(map[string]float64, len(accepted))
	acceptList := make([]string, len(accepted))
	for i, accept := range accepted {
		acceptParts := strings.SplitN(strings.TrimSpace(accept), ";", 2)
		q := 1.0
		if len(acceptParts) == 2 {
			acceptParts2 := strings.SplitN(acceptParts[1], "=", 2)
			if acceptParts2[0] == "q" && len(acceptParts2) == 2 {
				var err error
				q, err = strconv.ParseFloat(acceptParts2[1], 64)
				if err != nil {
					return nil, h.defaultResponseType, 400
				}
			} else {
				return nil, h.defaultResponseType, 400
			}
		}
		acceptMap[acceptParts[0]] = q
		acceptList[i] = acceptParts[0]
	}
	sort.SliceStable(acceptList, func(i, j int) bool {
		return acceptMap[acceptList[i]] > acceptMap[acceptList[j]]
	})

	for _, a := range acceptList {
		for _, marshaller := range h.responseMarshallers {
			if marshaller.SupportsMIME(a) {
				return marshaller, a, 200
			}
		}
	}
	return nil, h.defaultResponseType, 406
}

type internalRequest struct {
	writer  goHttp.ResponseWriter
	request *goHttp.Request
}

func (i *internalRequest) Decode(target interface{}) error {
	bytes, err := ioutil.ReadAll(i.request.Body)
	if err != nil {
		return &badRequestResponse
	}
	return json.Unmarshal(bytes, target)
}
