package auth

type PasswordRequest struct {
	User          string `json:"user"`
	RemoteAddress string `json:"remoteAddress"`
	SessionId     string `json:"sessionIdBase64"`
	Password      string `json:"passwordBase64"`
}

type PublicKeyRequest struct {
	User          string `json:"user"`
	RemoteAddress string `json:"remoteAddress"`
	SessionId     string `json:"sessionIdBase64"`
	// serialized key data in SSH wire format
	PublicKey string `json:"publicKeyBase64"`
}

type Response struct {
	Success bool `json:"success"`
}
