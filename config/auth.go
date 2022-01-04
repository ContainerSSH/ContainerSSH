package config

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"time"
)

// AuthConfig is the configuration of the authentication client.
type AuthConfig struct {
	// Method is the authentication method in use.
	Method AuthMethod `json:"method" yaml:"method" default:"webhook"`

	// Webhook is the configuration for webhook authentication that calls out to an external HTTP server.
	Webhook AuthWebhookClientConfig `json:"webhook" yaml:"webhook"`

	// OAuth2 is the configuration for OAuth2 authentication via Keyboard-Interactive.
	OAuth2 AuthOAuth2ClientConfig `json:"oauth2" yaml:"oauth2"`

	// Kerberos is the configuration for Kerberos authentication via GSSAPI and/or password.
	Kerberos AuthKerberosClientConfig `json:"kerberos" yaml:"kerberos"`

	// Authz is the authorization configuration. The authorization server
	// will receive a webhook after successful user authentication to
	// determine whether the specified user has access to the service.
	Authz AuthzConfig `json:"authz" yaml:"authz"`

	// AuthTimeout is the timeout for the overall authentication call (e.g. verifying a password). If the server
	// responds with a non-200 response the call will be retried until this timeout is reached. This timeout
	// should be increased to ~180s for OAuth2 login.
	AuthTimeout time.Duration `json:"authTimeout" yaml:"authTimeout" default:"60s"`

	// Deprecated: use the configuration in Webhook instead.
	HTTPClientConfiguration `json:",inline" yaml:",inline"`

	// Password is a flag to enable password authentication.
	// Deprecated: use the configuration in Webhook instead.
	Password bool `json:"password" yaml:"password" comment:"Perform password authentication" default:"true"`
	// PubKey is a flag to enable public key authentication.
	// Deprecated: use the configuration in Webhook instead.
	PubKey bool `json:"pubkey" yaml:"pubkey" comment:"Perform public key authentication" default:"true"`
}

func (c *AuthConfig) Validate() error {
	if err := c.Method.Validate(); err != nil {
		return fmt.Errorf("invalid method (%w)", err)
	}

	if c.Method == AuthMethodWebhook && c.URL != "" {
		//goland:noinspection GoDeprecation
		if c.Webhook.URL != "" {
			return fmt.Errorf("both auth.url and auth.webhook.url are set")
		}
		//goland:noinspection GoDeprecation
		c.Webhook.HTTPClientConfiguration = c.HTTPClientConfiguration
		//goland:noinspection GoDeprecation
		c.HTTPClientConfiguration = HTTPClientConfiguration{}
	}
	var err error
	switch c.Method {
	case AuthMethodWebhook:
		err = c.Webhook.Validate()
	case AuthMethodOAuth2:
		err = c.OAuth2.Validate()
	case AuthMethodKerberos:
		err = c.Kerberos.Validate()
	default:
		return fmt.Errorf("invalid method: %s", c.Method)
	}
	if err != nil {
		return fmt.Errorf("invalid %s client configuration (%w)", c.Method, err)
	}

	err = c.Authz.Validate()
	if err != nil {
		return fmt.Errorf("Invalid authz configuration (%w)", err)
	}

	return nil
}

type AuthMethod string

// Validate checks if the provided method is valid or not.
func (m AuthMethod) Validate() error {
	if m == "webhook" || m == "oauth2" || m == "kerberos" {
		return nil
	}
	return fmt.Errorf("invalid value for method: %s", m)
}

// AuthMethodWebhook authenticates using HTTP.
const AuthMethodWebhook AuthMethod = "webhook"

// AuthMethodOAuth2 authenticates by sending the user to a web interface using the keyboard-interactive facility.
const AuthMethodOAuth2 AuthMethod = "oauth2"

const AuthMethodKerberos AuthMethod = "kerberos"

// AuthWebhookClientConfig is the configuration for webhook authentication.
type AuthWebhookClientConfig struct {
	HTTPClientConfiguration `json:",inline" yaml:",inline"`

	// Password is a flag to enable password authentication.
	Password bool `json:"password" yaml:"password" comment:"Perform password authentication" default:"true"`
	// PubKey is a flag to enable public key authentication.
	PubKey bool `json:"pubkey" yaml:"pubkey" comment:"Perform public key authentication" default:"true"`
}

// Validate validates the authentication client configuration.
func (c *AuthWebhookClientConfig) Validate() error {
	if c.Timeout < 100*time.Millisecond {
		return fmt.Errorf("auth timeout value %s is too low, must be at least 100ms", c.Timeout.String())
	}
	if err := c.HTTPClientConfiguration.Validate(); err != nil {
		return err
	}
	return nil
}

// AuthOAuth2ClientConfig is the configuration for OAuth2-based authentication.
type AuthOAuth2ClientConfig struct {
	// Redirect is the configuration for the redirect URI server.
	Redirect OAuth2RedirectConfig `json:"redirect" yaml:"redirect"`

	// ClientID is the public identifier for the ContainerSSH installation.
	ClientID string `json:"clientId" yaml:"clientId"`
	// ClientSecret is the OAuth2 secret value needed to obtain the access token.
	ClientSecret string `json:"clientSecret" yaml:"clientSecret"`

	// Provider is the provider to use for authentication
	Provider OAuth2ProviderName `json:"provider" yaml:"provider"`

	// GitHub is the configuration for the GitHub provider.
	GitHub AuthGitHubConfig `json:"github" yaml:"github"`

	// OIDC is the configuration for the OpenID Connect provider.
	OIDC AuthOIDCConfig `json:"oidc" yaml:"oidc"`

	// QRCodeClients contains a list of strings that are used to identify SSH clients that can display an ASCII QR Code
	// for the device authorization flow. Each string is compiled as regular expressions and are used to match against
	// the client version string.
	//
	// This is done primarily because OpenSSH cuts off the sent text and mangles the drawing characters, so it cannot
	// be used to display a QR code.
	QRCodeClients []string `json:"qrCodeClients" yaml:"qrCodeClients"`

	// DeviceFlowClients is a list of clients that can use the device flow without sending keyboard-interactive
	// questions.
	DeviceFlowClients []string `json:"deviceFlowClients" yaml:"deviceFlowClients"`
}

// Validate validates if the OAuth2 client configuration is valid.
func (o *AuthOAuth2ClientConfig) Validate() error {
	if err := o.Redirect.Validate(); err != nil {
		return fmt.Errorf("invalid redirect configuration (%w)", err)
	}
	if o.ClientID == "" {
		return fmt.Errorf("empty client ID")
	}

	if o.ClientSecret == "" {
		return fmt.Errorf("empty client secret")
	}

	if err := o.Provider.Validate(); err != nil {
		return err
	}

	switch o.Provider {
	case AuthOAuth2GitHubProvider:
		if err := o.GitHub.Validate(); err != nil {
			return fmt.Errorf("invalid GitHub configuration (%w)", err)
		}
	case AuthOAuth2OIDCProvider:
		if err := o.OIDC.Validate(); err != nil {
			return fmt.Errorf("invalid OIDC configuration (%w)", err)
		}
	}

	return nil
}

type OAuth2ProviderName string

const (
	AuthOAuth2GitHubProvider OAuth2ProviderName = "github"
	AuthOAuth2OIDCProvider   OAuth2ProviderName = "oidc"
)

func (o OAuth2ProviderName) Validate() error {
	switch o {
	case AuthOAuth2GitHubProvider:
		return nil
	case AuthOAuth2OIDCProvider:
		return nil
	default:
		return fmt.Errorf("invalid Oauth2 provider")
	}
}

// OAuth2RedirectConfig is the configuration for the HTTP server that serves the page presented to the user after they
// are authenticated.
type OAuth2RedirectConfig struct {
	HTTPServerConfiguration `json:",inline" yaml:",inline"`

	// Webroot is a directory which contains all files that should be served as part of the return page
	// the user lands on when completing the OAuth2 authentication process. The webroot must contain an
	// index.html file, which will be served on the root URL. The files are read for each request and are not cached. If
	// the webroot is left empty the default ContainerSSH return page is presented.
	Webroot string `json:"webroot" yaml:"webroot"`
}

// Validate checks if the redirect server configuration is valid. Particularly, it checks the HTTP server
// parameters as well as if the webroot is valid and contains an index.html.
func (o OAuth2RedirectConfig) Validate() error {
	if err := o.HTTPServerConfiguration.Validate(); err != nil {
		return err
	}
	if o.Webroot != "" {
		webrootStat, err := os.Stat(o.Webroot)
		if err != nil {
			return fmt.Errorf("invalid webroot (%w)", err)
		}
		if !webrootStat.IsDir() {
			return fmt.Errorf("invalid webroot (not a directory)")
		}
		indexStat, err := os.Stat(path.Join(o.Webroot, "index.html"))
		if err != nil {
			return fmt.Errorf("webroot does not contain an index.html file (%w)", err)
		}
		if indexStat.IsDir() {
			return fmt.Errorf("webroot does not contain an index.html file (index.html is a directory)")
		}
	}
	return nil
}

// AuthGitHubConfig is the configuration structure for GitHub authentication.
//goland:noinspection GoVetStructTag
type AuthGitHubConfig struct {
	// URL is the base GitHub URL. Change this for GitHub Enterprise.
	URL string `json:"url" yaml:"url" default:"https://github.com"`

	// APIURL is the GitHub API URL. Change this for GitHub Enterprise.
	APIURL string `json:"apiurl" yaml:"apiurl" default:"https://api.github.com"`

	GitHubCACert `json:",inline" yaml:",inline"`

	// ClientCert is a PEM containing an x509 certificate to present to the server or a file name containing the PEM.
	ClientCert string `json:"cert" yaml:"cert" comment:"Client certificate file in PEM format."`

	// ClientKey is a PEM containing a private key to use to connect the server or a file name containing the PEM.
	ClientKey string `json:"key" yaml:"key" comment:"Client key file in PEM format."`

	// TLSVersion is the minimum TLS version to use.
	TLSVersion TLSVersion `json:"tlsVersion" yaml:"tlsVersion" default:"1.3"`

	// ECDHCurves is the list of curve algorithms to support.
	ECDHCurves ECDHCurveList `json:"curves" yaml:"curves" default:"[\"x25519\",\"secp256r1\",\"secp384r1\",\"secp521r1\"]"`

	// CipherSuites is a list of supported cipher suites.
	CipherSuites CipherSuiteList `json:"cipher" yaml:"cipher" default:"[\"TLS_AES_128_GCM_SHA256\",\"TLS_AES_256_GCM_SHA384\",\"TLS_CHACHA20_POLY1305_SHA256\",\"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256\",\"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256\"]"`

	// EnforceUsername requires that the GitHub username and the entered SSH username match. If this is set to false
	// the configuration server has to handle the GITHUB_USER connection parameter in order to obtain the correct
	// username as the SSH username cannot be trusted.
	EnforceUsername bool `json:"enforceUsername" yaml:"enforceUsername" default:"true"`
	// RequireOrgMembership checks if the user is part of the specified organization on GitHub. This requires the
	// read:org scope to be granted by the user. If the user does not grant this scope the authentication fails.
	RequireOrgMembership string `json:"requireOrgMembership" yaml:"requireOrgMembership"`
	// Require2FA requires the user to have two factor authentication enabled when logging in to this server. This
	// requires the read:user scope to be granted by the user. If the user does not grant this scope the authentication
	// fails.
	Require2FA bool `json:"require2FA" yaml:"require2FA"`

	// ExtraScopes asks the user to grant extra scopes to ContainerSSH. This is useful when the configuration server
	// needs these scopes to operate.
	ExtraScopes []string `json:"extraScopes" yaml:"extraScopes"`
	// EnforceScopes rejects the user authentication if the user fails to grant the scopes requested in extraScopes.
	EnforceScopes bool `json:"enforceScopes" yaml:"enforceScopes"`

	// RequestTimeout is the timeout for individual HTTP requests.
	RequestTimeout time.Duration `json:"timeout" yaml:"timeout" default:"10s"`
}

func (c *AuthGitHubConfig) Validate() (err error) {
	if _, err = url.Parse(c.URL); err != nil {
		return fmt.Errorf("invalid GitHub URL: %s (%w)", c.URL, err)
	}

	return nil
}

type AuthOIDCConfig struct {
	HTTPClientConfiguration `json:",inline" yaml:",inline"`

	// DeviceFlow enables or disables using the OIDC device flow.
	DeviceFlow bool `json:"deviceFlow" yaml:"deviceFlow" default:"true"`
	// AuthorizationCodeFlow enables or disables the OIDC authorization code flow.
	AuthorizationCodeFlow bool `json:"authorizationCodeFlow" yaml:"authorizationCodeFlow" default:"true"`
}

func (o *AuthOIDCConfig) Validate() error {
	if !o.DeviceFlow && !o.AuthorizationCodeFlow {
		return fmt.Errorf("at least one of deviceFlow or authorizationCodeFlow must be enabled")
	}
	return o.HTTPClientConfiguration.Validate()
}

// AuthzConfig is the configuration for the authorization flow
type AuthzConfig struct {
	HTTPClientConfiguration `json:",inline" yaml:",inline"`
	// Controls whether the authorization flow is enabled. If set to false
	// all authenticated users are allowed in the service.
	Enable bool `json:"enable" yaml:"enable" default:"false"`
}

func (k *AuthzConfig) Validate() error {
	if k.Enable {
		return k.HTTPClientConfiguration.Validate()
	}
	return nil
}

// AuthKerberosClientConfig is the configuration for the Kerberos authentication method.
type AuthKerberosClientConfig struct {
	// Keytab is the path to the kerberos keytab. If unset it defaults to
	// the default of /etc/krb5.keytab. If this file doesn't exist and
	// kerberos authentication is requested ContainerSSH will fail to start
	Keytab string `json:"keytab" yaml:"keytab" default:"/etc/krb5.keytab"`
	// Acceptor is the name of the keytab entry to authenticate against.
	// The value of this field needs to be in the form of `service/name`.
	//
	// The special value of `host` will authenticate clients only against
	// the service `host/hostname` where hostname is the system hostname
	// The special value of 'any' will authenticate against all keytab
	// entries regardless of name
	Acceptor string `json:"acceptor" yaml:"acceptor" default:"any"`
	// EnforceUsername specifies whether to check that the username of the
	// authenticated user matches the SSH username entered. If set to false
	// the authorization server must be responsible for ensuring proper
	// access control.
	//
	// WARNING: If authorization is unset and this is set to false all
	// authenticated users can log in to any account!
	EnforceUsername bool `json:"enforceUsername" yaml:"enforceUsername" default:"true"`
	// CredentialCachePath is the path in which the kerberos credentials
	// will be written inside the user containers.
	CredentialCachePath string `json:"credentialCachePath" yaml:"credentialCachePath" default:"/tmp/krb5cc"`
	// AllowPassword controls whether kerberos-based password
	// authentication should be allowed. If set to false only GSSAPI
	// authentication will be permitted
	AllowPassword bool `json:"allowPassword" yaml:"allowPassword" default:"true"`
	// ConfigPath is the path of the kerberos configuration file. This is
	// only used for password authentication.
	ConfigPath string `json:"configPath" yaml:"configPath" default:"/etc/containerssh/krb5.conf"`
	// ClockSkew is the maximum allowed clock skew for kerberos messages,
	// any messages older than this will be rejected. This value is also
	// used for the replay cache.
	ClockSkew time.Duration `json:"clockSkew" yaml:"clockSkew" default:"5m"`
}

func (k *AuthKerberosClientConfig) Validate() error {
	return nil
}
