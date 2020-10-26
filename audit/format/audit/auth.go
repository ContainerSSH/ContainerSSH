package audit

type PayloadAuthPassword struct {
	Username string `json:"username" yaml:"username"`
	Password []byte `json:"password" yaml:"password"`
}

type PayloadAuthPubKey struct {
	Username string `json:"username" yaml:"username"`
	Key      []byte `json:"key" yaml:"key"`
}
