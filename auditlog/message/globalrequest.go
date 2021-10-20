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
