package auth

type ConnectionMetadata struct {
	// Metadata is a set of key-value pairs that can be returned and
	// either consumed by the configuration server or exposed in the
	// backend as environment variables.
	Metadata    map[string]string `json:"metadata,omitempty"`
	// Environment is a set of key-value pairs that will be exposed to the
	// container as environment variables
	Environment map[string]string `json:"environment,omitempty"`
	// Files is a key-value pair of files to be placed inside containers.
	// The key represents the path to the file while the value is the
	// binary content.
	Files       map[string][]byte `json:"files,omitempty"`
}

// Transmit returns a copy of the Metadata containing only the metadata map for transmission to external servers (file and environment maps are considered sensitive by default)
func (m *ConnectionMetadata) Transmit() *ConnectionMetadata {
	if m == nil {
		return nil
	}
	return &ConnectionMetadata{
		Metadata: m.Metadata,
	}
}

// Merge merges a metadata object into the current one. In case of duplicated keys the one in the new struct take precedence
func (m *ConnectionMetadata) Merge(newmeta *ConnectionMetadata) {
	if m == nil {
		return
	}
	if newmeta == nil {
		return
	}
	for k, v := range newmeta.GetMetadata() {
		m.GetMetadata()[k] = v
	}
	for k, v := range newmeta.GetFiles() {
		m.GetFiles()[k] = v
	}
	for k, v := range newmeta.GetEnvironment() {
		m.GetEnvironment()[k] = v
	}
}

// GetMetadata returns an editable metadata map
func (m *ConnectionMetadata) GetMetadata() map[string]string {
	if m == nil {
		return nil
	}
	if m.Metadata == nil {
		m.Metadata = make(map[string]string)
	}
	return m.Metadata
}

// GetFiles returns an editable files map
func (m *ConnectionMetadata) GetFiles() map[string][]byte {
	if m == nil {
		return nil
	}
	if m.Files == nil {
		m.Files = make(map[string][]byte)
	}
	return m.Files
}

// GetFiles returns an editable files map
func (m *ConnectionMetadata) GetEnvironment() map[string]string {
	if m == nil {
		return nil
	}
	if m.Environment == nil {
		m.Environment = make(map[string]string)
	}
	return m.Environment
}
