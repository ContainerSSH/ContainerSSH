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

type PayloadRequestReverseForward struct {
	BindHost string
	BindPort uint32
}

func (p PayloadRequestReverseForward) Equals(other Payload) bool {
	p2, ok := other.(PayloadRequestReverseForward)
	if !ok {
		return false
	}
	return p.BindHost == p2.BindHost && p.BindPort == p2.BindPort
}

type PayloadNewForwardChannel struct {
	HostToConnect  string
	PortToConnect  uint32
	OriginatorHost string
	OriginatorPort uint32
}

func (p PayloadNewForwardChannel) Equals(other Payload) bool {
	p2, ok := other.(PayloadNewForwardChannel)
	if !ok {
		return false
	}
	return p.HostToConnect == p2.HostToConnect && p.PortToConnect == p2.PortToConnect && p.OriginatorHost == p2.OriginatorHost && p.OriginatorPort == p2.OriginatorPort
}

type PayloadNewReverseForwardChannel struct {
	ConnectedHost  string
	ConnectedPort  uint32
	OriginatorHost string
	OriginatorPort uint32
}

func (p PayloadNewReverseForwardChannel) Equals(other Payload) bool {
	p2, ok := other.(PayloadNewReverseForwardChannel)
	if !ok {
		return false
	}
	return p.ConnectedHost == p2.ConnectedHost && p.ConnectedPort == p2.ConnectedPort && p.OriginatorHost == p2.OriginatorHost && p.OriginatorPort == p2.OriginatorPort
}

type PayloadNewReverseX11ForwardChannel struct {
	OriginatorHost string
	OriginatorPort uint32
}

func (p PayloadNewReverseX11ForwardChannel) Equals(other Payload) bool {
	p2, ok := other.(PayloadNewReverseX11ForwardChannel)
	if !ok {
		return false
	}
	return p.OriginatorHost == p2.OriginatorHost && p.OriginatorPort == p2.OriginatorPort
}

type PayloadRequestStreamLocal struct {
	Path string
}

func (p PayloadRequestStreamLocal) Equals(other Payload) bool {
	p2, ok := other.(PayloadRequestStreamLocal)
	if !ok {
		return false
	}
	return p.Path == p2.Path
}
