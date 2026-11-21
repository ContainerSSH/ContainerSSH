package message

// PayloadRequestFailed is the payload for the TypeRequestFailed messages.
type PayloadRequestFailed struct {
	RequestID uint64 `json:"requestId" yaml:"reason"`
	Reason    string `json:"reason" yaml:"reason"`
}

// Equals compares two PayloadRequestFailed datasets.
func (p PayloadRequestFailed) Equals(other Payload) bool {
	p2, ok := other.(PayloadRequestFailed)
	if !ok {
		return false
	}
	return p.RequestID == p2.RequestID && p.Reason == p2.Reason
}
