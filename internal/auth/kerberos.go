package auth

// KerberosClient is the authenticator for Kerberos-based authentication. It supports both plain text and GSSAPI
// authentication.
type KerberosClient interface {
	PasswordAuthenticator
	GSSAPIAuthenticator
}
