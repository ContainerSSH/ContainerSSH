package message

type MetadataValue struct {
	Value     string `json:"value" yaml:"value"`
	Sensitive bool   `json:"sensitive" yaml:"sensitive"`
}

type MetadataBinaryValue struct {
	Value     []byte `json:"value" yaml:"value"`
	Sensitive bool   `json:"sensitive" yaml:"sensitive"`
}
