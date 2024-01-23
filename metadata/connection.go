package metadata

import (
	"encoding/json"
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"strings"
)

// RemoteAddress is an overlay for net.TCPAddr to provide JSON marshalling and unmarshalling.
//
//swagger:type string
type RemoteAddress net.TCPAddr

// String returns a string representation of this address.
func (r RemoteAddress) String() string {
	backend := net.TCPAddr(r)
	return backend.String()
}

// Network returns the network type ("tcp") of this address.
func (r RemoteAddress) Network() string {
	backend := net.TCPAddr(r)
	return backend.Network()
}

// AddrPort returns the netip.AddrPort component of this address.
func (r RemoteAddress) AddrPort() netip.AddrPort {
	backend := net.TCPAddr(r)
	return backend.AddrPort()
}

// MarshalJSON provides custom JSON marshalling to a string.
func (r RemoteAddress) MarshalJSON() ([]byte, error) {
	data, err := r.MarshalText()
	if err != nil {
		return nil, err
	}
	return json.Marshal(string(data))
}

// MarshalText provides custom marshalling to a string.
func (r RemoteAddress) MarshalText() ([]byte, error) {
	data := net.JoinHostPort(r.IP.String(), strconv.Itoa(r.Port))
	return []byte(data), nil
}

// UnmarshalJSON provides custom JSON unmarshalling from a string.
func (r *RemoteAddress) UnmarshalJSON(input []byte) error {
	var data string
	if err := json.Unmarshal(input, &data); err != nil {
		return fmt.Errorf("failed to unmarshal remote address data: %s (%w)", input, err)
	}
	return r.UnmarshalText([]byte(data))
}

// UnmarshalText provides custom unmarshalling from a string.
func (r *RemoteAddress) UnmarshalText(input []byte) error {
	parts := strings.Split(string(input), ":")
	if len(parts) < 2 {
		return fmt.Errorf("invalid IP:port combination: %s", input)
	}
	port, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		return fmt.Errorf("invalid port number: %d", port)
	}
	if port < 0 || port > 65535 {
		return fmt.Errorf("invalid port number: %d", port)
	}
	ip := strings.Join(parts[:len(parts)-1], ":")
	if ip[0] == '[' {
		ip = ip[1 : len(ip)-1]
	}
	parsedIP := net.ParseIP(ip)
	r.IP = parsedIP
	r.Port = port
	return nil
}

type AuthMethod string

const (
	AuthMethodPassword AuthMethod = "password"
	AuthMethodPubKey AuthMethod = "publickey"
	AuthMethodKeyboardInteractive AuthMethod = "keyboard-interactive"
)

// ConnectionMetadata holds a metadata structure passed around with a metadata. Its main purpose is to allow an
// authentication or authorization module to configure data exposed to the configuration server or the backend.
//
// swagger:model ConnectionMetadata
type ConnectionMetadata struct {
	// RemoteAddress is the IP address and port of the user trying to authenticate.
	//
	// required: true
	// in: body
	RemoteAddress RemoteAddress `json:"remoteAddress"`

	// ConnectionID is an opaque ID to identify the SSH connection in question.
	//
	// required: true
	// in: body
	ConnectionID string `json:"connectionId"`

	// AuthenticationMethods are the authentication methods that can be
	// used to authenticate this connection
	AuthenticationMethods map[AuthMethod]bool `json:"-"`

	// Metadata is a set of key-value pairs that carry additional information from the authentication and configuration
	// system to the backends. Backends can expose this information as container labels, environment variables, or
	// other places.
	//
	// required: false
	// in: body
	Metadata map[string]Value `json:"metadata,omitempty"`

	// Environment is a set of key-value pairs provided by the authentication or configuration system and may be
	// exposed by the backend.
	//
	// required: false
	// in: body
	Environment map[string]Value `json:"environment,omitempty"`

	// Files is a key-value pair of file names and their content set by the authentication or configuration system
	// and consumed by the backend.
	//
	// required: false
	// in: body
	Files map[string]BinaryValue `json:"files,omitempty"`
}

func (meta ConnectionMetadata) StartAuthentication(
	clientVersion string,
	username string,
) ConnectionAuthPendingMetadata {
	return ConnectionAuthPendingMetadata{
		meta,

		clientVersion,
		username,
	}
}

// NewTestMetadata provides a metadata set useful for testing.
func NewTestMetadata() ConnectionMetadata {
	return ConnectionMetadata{
		RemoteAddress: RemoteAddress(
			net.TCPAddr{
				IP:   net.ParseIP("127.0.0.1"),
				Port: 22,
			},
		),
		ConnectionID: "0123456789ABCDEF",
		Metadata:     map[string]Value{},
		Environment:  map[string]Value{},
		Files:        map[string]BinaryValue{},
	}
}

// ConnectionAuthPendingMetadata is a variant of ConnectionMetadata which is used when the client has already
// provided a Username, but the authentication has not completed yet.
//
// swagger:model ConnectionAuthPendingMetadata
type ConnectionAuthPendingMetadata struct {
	ConnectionMetadata `json:",inline"`

	// ClientVersion contains the version string the connecting client sent if any. May be empty if the client did not
	// provide a client version.
	//
	// required: false
	// in: body
	ClientVersion string `json:"clientVersion"`

	// Username is the username provided on login by the client. This may, but must not necessarily match the
	// authenticated username.
	//
	// required: true
	// in: body
	Username string `json:"username"`
}

func NewTestAuthenticatingMetadata(username string) ConnectionAuthPendingMetadata {
	return ConnectionAuthPendingMetadata{
		NewTestMetadata(),
		"SSH-2.0-FooSSH",
		username,
	}
}

// Authenticated creates a copy after authentication.
func (c ConnectionAuthPendingMetadata) Authenticated(username string) ConnectionAuthenticatedMetadata {
	return ConnectionAuthenticatedMetadata{
		c,
		username,
	}
}

// AuthFailed creates a copy after a failed authentication to be passed along with an authentication failure.
func (c ConnectionAuthPendingMetadata) AuthFailed() ConnectionAuthenticatedMetadata {
	return ConnectionAuthenticatedMetadata{
		c,
		"",
	}
}

// ConnectionAuthenticatedMetadata is a variant of ConnectionMetadata which is used once the authentication has been
// completed. It contains the AuthenticatedUsername provided by the authentication system.
type ConnectionAuthenticatedMetadata struct {
	ConnectionAuthPendingMetadata `json:",inline"`

	// AuthenticatedUsername contains the username that was actually verified. This may differ from LoginUsername when,
	// for example OAuth2 or Kerberos authentication is used. This field is empty until the authentication phase is
	// completed.
	//
	// required: false
	// in: body
	AuthenticatedUsername string `json:"authenticatedUsername,omitempty"`
}

// Merge merges the newMeta into the current metadata structure. If environment, files, or metadata are set, these
// override the current variable. If they are not set, the result is kept.
func (c *ConnectionAuthenticatedMetadata) Merge(newMeta ConnectionAuthenticatedMetadata) ConnectionAuthenticatedMetadata {
	c.ConnectionAuthPendingMetadata.ConnectionMetadata.Merge(newMeta.ConnectionAuthPendingMetadata.ConnectionMetadata)
	if newMeta.AuthenticatedUsername != "" {
		c.AuthenticatedUsername = newMeta.AuthenticatedUsername
	}
	return *c
}

func (meta ConnectionAuthenticatedMetadata) Channel(channelID uint64) ChannelMetadata {
	return ChannelMetadata{
		meta,
		channelID,
	}
}

// Merge merges a metadata object into the current one. If the newMeta contains environment, files, or metadata, they
// will override the current content.
func (meta *ConnectionMetadata) Merge(newMeta ConnectionMetadata) {
	if newMeta.GetMetadata() != nil {
		meta.Metadata = newMeta.GetMetadata()
	}
	if newMeta.GetFiles() != nil {
		meta.Files = newMeta.GetFiles()
	}
	if newMeta.GetEnvironment() != nil {
		meta.Environment = newMeta.GetEnvironment()
	}
}

// GetMetadata returns an editable metadata map.
func (meta *ConnectionMetadata) GetMetadata() map[string]Value {
	if meta == nil {
		return nil
	}
	if meta.Metadata == nil {
		meta.Metadata = make(map[string]Value)
	}
	return meta.Metadata
}

// GetFiles returns an editable files map.
func (meta *ConnectionMetadata) GetFiles() map[string]BinaryValue {
	if meta == nil {
		return nil
	}
	if meta.Files == nil {
		meta.Files = make(map[string]BinaryValue)
	}
	return meta.Files
}

// GetEnvironment returns an editable files map.
func (meta *ConnectionMetadata) GetEnvironment() map[string]Value {
	if meta == nil {
		return nil
	}
	if meta.Environment == nil {
		meta.Environment = make(map[string]Value)
	}
	return meta.Environment
}
