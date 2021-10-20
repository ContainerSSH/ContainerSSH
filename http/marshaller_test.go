package http

import (
	"reflect"
	"testing"
)

func TestTextMarshal(t *testing.T) {
	dataSet := []interface{}{
		42,
		int8(42),
		int16(42),
		int32(42),
		int64(42),
		uint(42),
		uint8(42),
		uint16(42),
		uint32(42),
		uint64(42),
		"42",
		testData{},
		&testData{},
	}

	marshaller := &textMarshaller{}
	for _, v := range dataSet {
		t.Run(reflect.TypeOf(v).Name(), func(t *testing.T) {
			result, err := marshaller.Marshal(v)
			if err != nil {
				t.Fatal(err)
			}
			if string(result) != "42" {
				t.Fatalf("unexpected marshal result: %s", result)
			}
		})
	}
}

type testData struct{}

func (t testData) MarshalText() string {
	return "42"
}
