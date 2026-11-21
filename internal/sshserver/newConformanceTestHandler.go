package sshserver

func newConformanceTestHandler(backend NetworkConnectionHandler) *conformanceTestHandler {
	return &conformanceTestHandler{backend: backend}
}
