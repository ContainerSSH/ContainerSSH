package message

//go:generate go run github.com/containerssh/libcontainerssh/cmd/generate-message-codes backend.go BACKEND.md "Backend"

// EBackendConfig indicates that there is an error in the backend configuration.
const EBackendConfig = "BACKEND_CONFIG_ERROR"
