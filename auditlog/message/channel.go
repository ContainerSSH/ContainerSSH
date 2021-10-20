package message

// PayloadNewChannel is a payload that signals a request for a new SSH channel
type PayloadNewChannel struct {
	ChannelType string `json:"channelType" yaml:"channelType"`
}

// Equals compares two PayloadNewChannel payloads.
func (p PayloadNewChannel) Equals(other Payload) bool {
	p2, ok := other.(PayloadNewChannel)
	if !ok {
		return false
	}
	return p.ChannelType == p2.ChannelType
}

// PayloadNewChannelFailed is a payload that signals that a request for a new channel has failed.
type PayloadNewChannelFailed struct {
	ChannelType string `json:"channelType" yaml:"channelType"`
	Reason      string `json:"reason" yaml:"reason"`
}

// Equals compares two PayloadNewChannelFailed payloads.
func (p PayloadNewChannelFailed) Equals(other Payload) bool {
	p2, ok := other.(PayloadNewChannelFailed)
	if !ok {
		return false
	}
	return p.ChannelType == p2.ChannelType && p.Reason == p2.Reason
}

// PayloadNewChannelSuccessful is a payload that signals that a channel request was successful.
type PayloadNewChannelSuccessful struct {
	ChannelType string `json:"channelType" yaml:"channelType"`
}

// Equals compares two PayloadNewChannelSuccessful payloads.
func (p PayloadNewChannelSuccessful) Equals(other Payload) bool {
	p2, ok := other.(PayloadNewChannelSuccessful)
	if !ok {
		return false
	}
	return p.ChannelType == p2.ChannelType
}
