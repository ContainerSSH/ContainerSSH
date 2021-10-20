package message

// PayloadConnect is the payload for TypeConnect messages.
type PayloadConnect struct {
	RemoteAddr string `json:"remoteAddr" yaml:"remoteAddr"` // RemoteAddr contains the IP address of the connecting user.
	Country    string `json:"country" yaml:"country"`       // Country contains the country code looked up from the IP address. Contains "XX" if the lookup failed.
}

// Equals compares two PayloadConnect datasets.
func (p PayloadConnect) Equals(other Payload) bool {
	p2, ok := other.(PayloadConnect)
	if !ok {
		return false
	}
	return p.RemoteAddr == p2.RemoteAddr
}
