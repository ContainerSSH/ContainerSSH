package http

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// responseMarshaller is an interface to cover all encoders for HTTP response bodies.
type responseMarshaller interface {
	SupportsMIME(mime string) bool
	Marshal(body interface{}) ([]byte, error)
}

//region JSON

type jsonMarshaller struct {
}

func (j *jsonMarshaller) SupportsMIME(mime string) bool {
	return mime == "application/json" || mime == "application/*" || mime == "*/*"
}

func (j *jsonMarshaller) Marshal(body interface{}) ([]byte, error) {
	return json.Marshal(body)
}

func (j *jsonMarshaller) Unmarshal(body []byte, target interface{}) error {
	return json.Unmarshal(body, target)
}

// endregion

// region Text
type TextMarshallable interface {
	MarshalText() string
}

type textMarshaller struct {
}

func (t *textMarshaller) SupportsMIME(mime string) bool {
	// HTML output might be better suited to piping through a templating engine.
	return mime == "text/html" || mime == "text/plain" || mime == "text/*" || mime == "*/*"
}

func (t *textMarshaller) Marshal(body interface{}) ([]byte, error) {
	switch assertedBody := body.(type) {
	case TextMarshallable:
		return []byte(assertedBody.MarshalText()), nil
	case string:
		return []byte(assertedBody), nil
	case int:
		return t.marshalNumber(body)
	case int8:
		return t.marshalNumber(body)
	case int16:
		return t.marshalNumber(body)
	case int32:
		return t.marshalNumber(body)
	case int64:
		return t.marshalNumber(body)
	case uint:
		return t.marshalNumber(body)
	case uint8:
		return t.marshalNumber(body)
	case uint16:
		return t.marshalNumber(body)
	case uint32:
		return t.marshalNumber(body)
	case uint64:
		return t.marshalNumber(body)
	case bool:
		if body.(bool) {
			return []byte("true"), nil
		} else {
			return []byte("false"), nil
		}
	case uintptr:
		return t.marshalPointer(body)
	default:
		return nil, fmt.Errorf("cannot marshal unknown type: %v", body)
	}
}

func (t *textMarshaller) marshalNumber(body interface{}) ([]byte, error) {
	return []byte(fmt.Sprintf("%d", body)), nil
}

func (t *textMarshaller) marshalPointer(body interface{}) ([]byte, error) {
	ptr := body.(uintptr)
	return t.Marshal(reflect.ValueOf(ptr).Elem())
}

// endregion
