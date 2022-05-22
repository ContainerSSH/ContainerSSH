package metadata_test

import (
	"encoding/json"
	"net"
	"testing"

	"github.com/containerssh/libcontainerssh/metadata"
)

func TestMarshalRemoteAddress(t *testing.T) {
	data := []metadata.RemoteAddress{
		{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 1234,
		},
		{
			IP: net.ParseIP("fe80::"),
		},
		{
			IP:   net.ParseIP("fe80::1"),
			Port: 2222,
		},
	}
	for _, entry := range data {
		marshalled, err := json.Marshal(entry)
		if err != nil {
			t.Fatalf("Failed to JSON marshal %s:%d (%v).", entry.IP, entry.Port, err)
		}
		var unmarshalled metadata.RemoteAddress
		if err := json.Unmarshal(marshalled, &unmarshalled); err != nil {
			t.Fatalf("Failed to JSON unmarshal %s (%v).", marshalled, err)
		}
		if entry.IP.String() != unmarshalled.IP.String() {
			t.Fatalf(
				"Mismatching IP after unmarshal, expected: %s, got: %s",
				entry.IP.String(),
				unmarshalled.IP.String(),
			)
		}
		if entry.Port != unmarshalled.Port {
			t.Fatalf("Mismatching port after unmarshal, expected: %d, got: %d", entry.Port, unmarshalled.Port)
		}
	}
}

func TestMarshalConnectionMetadata(t *testing.T) {
	meta := metadata.ConnectionMetadata{
		RemoteAddress: metadata.RemoteAddress{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 2222,
		},
		ConnectionID: "asdf",
		Metadata: map[string]metadata.Value{
			"test": {
				Value:     "testing",
				Sensitive: true,
			},
		},
		Environment: nil,
		Files:       nil,
	}

	data, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("Failed to marshal test data (%v).", err)
	}
	var newMeta metadata.ConnectionMetadata
	if err := json.Unmarshal(data, &newMeta); err != nil {
		t.Fatalf("Failed to unmarshal test data (%v):\n%s", err, data)
	}
	if newMeta.RemoteAddress.IP.String() != meta.RemoteAddress.IP.String() {
		t.Fatalf(
			"Mismatched IP address (expected: %s, got: %s).",
			meta.RemoteAddress.IP.String(),
			newMeta.RemoteAddress.IP.String(),
		)
	}
	if newMeta.RemoteAddress.Port != meta.RemoteAddress.Port {
		t.Fatalf(
			"Mismatched port (expected: %d, got: %d).",
			meta.RemoteAddress.Port,
			newMeta.RemoteAddress.Port,
		)
	}
	if newMeta.Metadata["test"].Value != meta.Metadata["test"].Value {
		t.Fatalf(
			"Mismatched metadata entry value (expected: %s, got: %s).",
			meta.Metadata["test"].Value,
			newMeta.Metadata["test"].Value,
		)
	}
	if newMeta.Metadata["test"].Sensitive != meta.Metadata["test"].Sensitive {
		t.Fatalf(
			"Mismatched metadata entry sensitive field (expected: %t, got: %t).",
			meta.Metadata["test"].Sensitive,
			newMeta.Metadata["test"].Sensitive,
		)
	}
}
