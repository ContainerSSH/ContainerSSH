package metadata

type ChannelMetadata struct {
	Connection ConnectionAuthenticatedMetadata `json:"connection"`

	// ChannelID signals the unique number of the channel within the connection.
	ChannelID uint64 `json:"channelID"`
}
