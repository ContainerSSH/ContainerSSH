package sshserver

import (
	"encoding/hex"

	"github.com/google/uuid"
)

// GenerateConnectionID generates a globally unique connection ID consisting of hexadecimal characters.
func GenerateConnectionID() string {
	connectionIDBinary, err := uuid.New().MarshalBinary()
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(connectionIDBinary)
}
