package message

// PayloadGlobalRequestUnknown Is a payload for the TypeGlobalRequestUnknown messages.
type PayloadGlobalRequestUnknown struct {
	RequestType string `json:"requestType" yaml:"requestType"`
}

// Equals Compares two PayloadGlobalRequestUnknown payloads.
func (p PayloadGlobalRequestUnknown) Equals(other Payload) bool {
	p2, ok := other.(PayloadGlobalRequestUnknown)
	if !ok {
		return false
	}
	return p.RequestType == p2.RequestType
}

// PayloadGlobalRequestDecodeFailed is a payload that signals a supported request that the server was unable to decode
type PayloadGlobalRequestDecodeFailed struct {
	RequestID uint64 `json:"requestId" yaml:"requestId"`

	RequestType string `json:"requestType" yaml:"requestType"`
	Payload     []byte `json:"payload" yaml:"payload"`
	Reason      string `json:"reason" yaml:"reason"`
}

func (p PayloadGlobalRequestDecodeFailed) Equals(other Payload) bool {
	p2, ok := other.(PayloadGlobalRequestDecodeFailed)
	if !ok {
		return false
	}
	return p.RequestID == p2.RequestID && p.RequestType == p2.RequestType && p.Reason == p2.Reason
}