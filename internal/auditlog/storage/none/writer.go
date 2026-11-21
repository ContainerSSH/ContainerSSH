package none

type nullWriteCloser struct {
}

func (w *nullWriteCloser) SetMetadata(_ int64, _ string, _ string, _ *string) {
}

func (w *nullWriteCloser) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (w *nullWriteCloser) Close() error {
	return nil
}
