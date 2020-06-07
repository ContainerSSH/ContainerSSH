package protocol

type PasswordAuthRequest struct {
	User          string `json:"user"`
	RemoteAddress string `json:"remoteAddress"`
	SessionId     string `json:"sessionId"`
	Password      string `json:"passwordBase64"`
}

type PublicKeyAuthRequest struct {
	User          string `json:"user"`
	RemoteAddress string `json:"remoteAddress"`
	SessionId     string `json:"sessionId"`
	// serialized key data in SSH wire format
	PublicKey string `json:"publicKeyBase64"`
}

type AuthResponse struct {
	Success bool `json:"success"`
}
