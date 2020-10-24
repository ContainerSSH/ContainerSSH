package protocol

type PayloadNewChannel struct {
	ChannelType string `json:"channelType" yaml:"channelType"`
}

type PayloadNewChannelFailed struct {
	ChannelType string `json:"channelType" yaml:"channelType"`
	Reason      string `json:"reason" yaml:"reason"`
}

type PayloadNewChannelSuccessful struct {
	ChannelType string `json:"channelType" yaml:"channelType"`
}
