// +build linux freebsd openbsd darwin

package auth

// GitHubCACert contains the CA certificate on non-Windows platforms. This is embedded into the main
// configuration structure as a workaround for Go bug 16736.
type GitHubCACert struct {
	// CACert is the PEM-encoded CA certificate, or file containing a PEM-encoded CA certificate used to verify the
	// GitHub server certificate.
	CACert string `json:"cacert" yaml:"cacert" default:""`
}
