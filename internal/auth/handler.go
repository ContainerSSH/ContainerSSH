package auth

type Handler interface {
	// OnPassword is called if the client requests a password authentication.
	//
	// - Username is the username the user entered.
	// - Password is the password the user entered.
	// - RemoteAddress is the IP address of the user.
	// - SessionID is an opaque identifier for the current session.
	//
	// The method must return a boolean if the authentication was successful, and an error if the authentication failed
	// for other reasons (e.g. backend database was not available). If an error is returned the server responds with
	// a HTTP 500 response.
	OnPassword(
		Username string,
		Password []byte,
		RemoteAddress string,
		ConnectionID string,
	) (bool, map[string]string, error)

	// OnPubKey is called when the client requests a public key authentication.
	//
	// - Username is the username the user entered.
	// - PublicKey is the public key of the user in the authorized key format.
	// - RemoteAddress is the IP address of the user.
	// - SessionID is an opaque identifier for the current session.
	//
	// The method must return a boolean if the authentication was successful, and an error if the authentication failed
	// for other reasons (e.g. backend database was not available). If an error is returned the server responds with
	// a HTTP 500 response.
	OnPubKey(
		Username string,
		PublicKey string,
		RemoteAddress string,
		ConnectionID string,
	) (bool, map[string]string, error)
}
