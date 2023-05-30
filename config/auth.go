package config

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path"
	"time"
)

// AuthConfig is the configuration of the authentication client.
type AuthConfig struct {
	// PasswordAuth configures how users are authenticated with passwords. If not set, password authentication is
	// disabled.
	PasswordAuth PasswordAuthConfig `json:"password" yaml:"password"`

	// PublicKeyAuth configures how users are authenticated with their public keys. If not set, public key
	// authentication is disabled.
	PublicKeyAuth PublicKeyAuthConfig `json:"publicKey" yaml:"publicKey"`

	// KeyboardInteractiveAuth configures how users are authenticated using the keyboard-interactive (question-answer)
	// method. If this is empty, keyboard-interactive is disabled.
	KeyboardInteractiveAuth KeyboardInteractiveAuthConfig `json:"keyboardInteractive" yaml:"keyboardInteractive"`

	// GSSAPIAuth configures how users are authenticated using the GSSAPI (typically Kerberos) method. If this is empty,
	// GSSAPI authentication is disabled.
	GSSAPIAuth GSSAPIAuthConfig `json:"gssapi" yaml:"gssapi"`

	// Authz is the authorization configuration. The authorization server will receive a webhook after successful user
	// authentication to determine whether the specified user has access to the service. If not set authorization is
	// disabled. It is strongly recommended you configure AuthZ in case of oAuth2 and GSSAPI methods as these methods
	// do not verify the provided SSH username.
	Authz AuthzConfig `json:"authz" yaml:"authz"`

	// AuthTimeout is the timeout for the overall authentication call (e.g. verifying a password). If the server
	// responds with a non-200 response the call will be retried until this timeout is reached. This timeout
	// should be increased to ~180s for OAuth2 login.
	// Deprecated: please use the individual authentication methods instead.
	AuthTimeout time.Duration `json:"authTimeout" yaml:"authTimeout" default:"60s"`

	// Deprecated: please use the individual authentication configurations instead.
	HTTPClientConfiguration `json:",inline" yaml:",inline"`

	// Password is a flag to enable password authentication.
	// Deprecated: use PasswordAuth instead.
	Password *bool `json:"-" yaml:"-" comment:"Perform password authentication" default:"true"`
	// PubKey is a flag to enable public key authentication.
	// Deprecated: use PublicKeyAuth instead.
	PubKey *bool `json:"pubkey" yaml:"pubkey" comment:"Perform public key authentication" default:"true"`
}

type legacyAuthConfig struct {
	PasswordAuth            PasswordAuthConfig            `json:"-" yaml:"-"`
	PublicKeyAuth           PublicKeyAuthConfig           `json:"publicKey" yaml:"publicKey"`
	KeyboardInteractiveAuth KeyboardInteractiveAuthConfig `json:"keyboardInteractive" yaml:"keyboardInteractive"`
	GSSAPIAuth              GSSAPIAuthConfig              `json:"gssapi" yaml:"gssapi"`
	Authz                   AuthzConfig                   `json:"authz" yaml:"authz"`

	HTTPClientConfiguration `json:",inline" yaml:",inline"`
	AuthTimeout             time.Duration `json:"authTimeout" yaml:"authTimeout" default:"60s"`
	Password                *bool         `json:"password,omitempty" yaml:"password,omitempty"`
	PubKey                  *bool         `json:"pubkey,omitempty" yaml:"pubkey,omitempty"`
}

func (l *legacyAuthConfig) Set(c *AuthConfig) {
	c.PasswordAuth = l.PasswordAuth
	c.PublicKeyAuth = l.PublicKeyAuth
	c.KeyboardInteractiveAuth = l.KeyboardInteractiveAuth
	c.GSSAPIAuth = l.GSSAPIAuth
	c.Authz = l.Authz
	c.HTTPClientConfiguration = l.HTTPClientConfiguration
	c.Password = l.Password
	c.PubKey = l.PubKey
}

type newAuthConfig struct {
	PasswordAuth            PasswordAuthConfig            `json:"password" yaml:"password"`
	PublicKeyAuth           PublicKeyAuthConfig           `json:"publicKey" yaml:"publicKey"`
	KeyboardInteractiveAuth KeyboardInteractiveAuthConfig `json:"keyboardInteractive" yaml:"keyboardInteractive"`
	GSSAPIAuth              GSSAPIAuthConfig              `json:"gssapi" yaml:"gssapi"`
	Authz                   AuthzConfig                   `json:"authz" yaml:"authz"`

	HTTPClientConfiguration `json:",inline" yaml:",inline"`
	AuthTimeout             time.Duration `json:"authTimeout" yaml:"authTimeout"`
	Password                *bool         `json:"-" yaml:"-"`
	PubKey                  *bool         `json:"pubkey" yaml:"pubkey"`
}

func (n *newAuthConfig) Set(c *AuthConfig) {
	c.PasswordAuth = n.PasswordAuth
	c.PublicKeyAuth = n.PublicKeyAuth
	c.KeyboardInteractiveAuth = n.KeyboardInteractiveAuth
	c.GSSAPIAuth = n.GSSAPIAuth
	c.Authz = n.Authz
	c.HTTPClientConfiguration = n.HTTPClientConfiguration
	c.Password = n.Password
	c.PubKey = n.PubKey
}

func (c *AuthConfig) UnmarshalJSON(data []byte) error {
	legacyConfig := &legacyAuthConfig{}
	if err := json.Unmarshal(data, legacyConfig); err == nil {
		legacyConfig.Set(c)
		return nil
	}
	newConfig := &newAuthConfig{}
	if err := json.Unmarshal(data, newConfig); err != nil {
		return err
	}
	newConfig.Set(c)
	return nil
}

func (c *AuthConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	legacyConfig := &legacyAuthConfig{}
	if err := unmarshal(legacyConfig); err == nil {
		legacyConfig.Set(c)
		return nil
	}
	newConfig := &newAuthConfig{}
	if err := unmarshal(newConfig); err != nil {
		return err
	}
	newConfig.Set(c)
	return nil
}

// Validate checks if the authentication configuration is valid.
func (c *AuthConfig) Validate() error {
	if //goland:noinspection GoDeprecation
	c.PasswordAuth.Method == PasswordAuthMethodDisabled &&
		c.PublicKeyAuth.Method == PubKeyAuthMethodDisabled &&
		c.KeyboardInteractiveAuth.Method == KeyboardInteractiveAuthMethodDisabled &&
		c.GSSAPIAuth.Method == GSSAPIAuthMethodDisabled &&
		(((c.Password == nil || !*c.Password) && (c.PubKey == nil || !*c.PubKey)) && c.URL == "") {
		return fmt.Errorf("no authentication method configured, please configure at least one")
	}
	if c.PasswordAuth.Method != PasswordAuthMethodDisabled {
		//goland:noinspection GoDeprecation
		if c.URL != "" && c.Password != nil && *c.Password {
			return newError(
				"url",
				"both the password authentication and the legacy url have been provided, please use the new configuration format",
			)
		}
		if err := c.PasswordAuth.Validate(); err != nil {
			return wrap(err, "password")
		}
	}
	if c.PublicKeyAuth.Method != PubKeyAuthMethodDisabled {
		//goland:noinspection GoDeprecation
		if c.URL != "" && c.PubKey != nil && *c.PubKey {
			return newError(
				"url",
				"both the pubkey authentication and the legacy url have been provided, please use the new configuration format",
			)
		}
		if err := c.PublicKeyAuth.Validate(); err != nil {
			return wrap(err, "publicKey")
		}
	}
	if c.KeyboardInteractiveAuth.Method != KeyboardInteractiveAuthMethodDisabled {
		if err := c.KeyboardInteractiveAuth.Validate(); err != nil {
			return wrap(err, "keyboardInteractive")
		}
	}
	if c.GSSAPIAuth.Method != GSSAPIAuthMethodDisabled {
		if err := c.GSSAPIAuth.Validate(); err != nil {
			return wrap(err, "gssapi")
		}
	}
	if c.Authz.Method != AuthzMethodDisabled {
		if err := c.Authz.Validate(); err != nil {
			return wrap(err, "authz")
		}
	}
	//goland:noinspection GoDeprecation
	if ((c.Password != nil && *c.Password) || (c.PubKey != nil && *c.PubKey)) && c.URL != "" {
		//goland:noinspection GoDeprecation
		return c.HTTPClientConfiguration.Validate()
	}

	return nil
}

// region AuthMethod

// AuthMethod is a listing of all authentication methods. These methods are not guaranteed to support any particular
// authentication method.
type AuthMethod string

// Validate checks if the provided method is valid or not.
func (m AuthMethod) Validate() error {
	if m == "webhook" || m == "oauth2" || m == "kerberos" {
		return nil
	}
	return fmt.Errorf("invalid value for method: %s", m)
}

// AuthMethodDisabled disables the authentication method
const AuthMethodDisabled AuthMethod = ""

// AuthMethodWebhook authenticates using HTTP.
const AuthMethodWebhook AuthMethod = "webhook"

// AuthMethodOAuth2 authenticates by sending the user to a web interface using the keyboard-interactive facility.
const AuthMethodOAuth2 AuthMethod = "oauth2"

// AuthMethodKerberos authenticates using the Kerberos method.
const AuthMethodKerberos AuthMethod = "kerberos"

// endregion

// region PasswordAuth

// PasswordAuthConfig configures how password authentications are performed.
type PasswordAuthConfig struct {
	// Method is the authenticator to use for passwords.
	Method PasswordAuthMethod `json:"method" yaml:"method" default:""`

	// Webhook configures the webhook authenticator for password authentication.
	Webhook AuthWebhookClientConfig `json:"webhook" yaml:"webhook"`

	// Kerberos configures the Kerberos authenticator for password authentication.
	Kerberos AuthKerberosClientConfig `json:"kerberos" yaml:"kerberos"`
}

// Validate checks the password configuration structure for misconfiguration.
func (c PasswordAuthConfig) Validate() error {
	if err := c.Method.Validate(); err != nil {
		return fmt.Errorf("invalid method for password authentication (%w)", err)
	}
	switch c.Method {
	case PasswordAuthMethodDisabled:
		return nil
	case PasswordAuthMethodWebhook:
		return c.Webhook.Validate()
	case PasswordAuthMethodKerberos:
		return c.Kerberos.Validate()
	default:
		return fmt.Errorf("BUG: unsupported password authenticator: %s", c.Method)
	}
}

// PasswordAuthMethod provides the methods usable for password authentication.
type PasswordAuthMethod string

// Validate checks if the provided method is valid or not.
func (m PasswordAuthMethod) Validate() error {
	if m == PasswordAuthMethodDisabled || m == PasswordAuthMethodWebhook || m == PasswordAuthMethodKerberos {
		return nil
	}
	return fmt.Errorf("invalid value for method: %s", m)
}

// PasswordAuthMethodDisabled disables password authentication.
const PasswordAuthMethodDisabled PasswordAuthMethod = PasswordAuthMethod(AuthMethodDisabled)

// PasswordAuthMethodWebhook authenticates using an HTTP webhook.
const PasswordAuthMethodWebhook PasswordAuthMethod = PasswordAuthMethod(AuthMethodWebhook)

// PasswordAuthMethodKerberos authenticates passwords using Kerberos.
const PasswordAuthMethodKerberos PasswordAuthMethod = PasswordAuthMethod(AuthMethodKerberos)

// endregion

// region PubKeyAuth

// PublicKeyAuthConfig holds the configuration for public key authentication.
type PublicKeyAuthConfig struct {
	// Method is the authenticator to use for public keys.
	Method PublicKeyAuthMethod `json:"method" yaml:"method" default:""`

	// Webhook configures the webhook authenticator for public key authentication.
	Webhook AuthWebhookClientConfig `json:"webhook" yaml:"webhook"`
}

func (c PublicKeyAuthConfig) Validate() error {
	if err := c.Method.Validate(); err != nil {
		return fmt.Errorf("invalid public key authentication configuration (%w)", err)
	}
	switch c.Method {
	case PubKeyAuthMethodDisabled:
		return nil
	case PubKeyAuthMethodWebhook:
		return c.Webhook.Validate()
	default:
		return fmt.Errorf("BUG: unsupported public key authenticator: %s", c.Method)
	}
}

// PublicKeyAuthMethod provides the methods usable for public key authentication.
type PublicKeyAuthMethod string

// Validate checks if the provided method is valid or not.
func (m PublicKeyAuthMethod) Validate() error {
	if m == PubKeyAuthMethodDisabled || m == PubKeyAuthMethodWebhook {
		return nil
	}
	return fmt.Errorf("invalid value for method: %s", m)
}

// PubKeyAuthMethodDisabled disables public key authentication.
const PubKeyAuthMethodDisabled PublicKeyAuthMethod = PublicKeyAuthMethod(AuthMethodDisabled)

// PubKeyAuthMethodWebhook authenticates using an HTTP webhook.
const PubKeyAuthMethodWebhook PublicKeyAuthMethod = PublicKeyAuthMethod(AuthMethodWebhook)

// endregion

// region Keyboard-interactive

// KeyboardInteractiveAuthConfig configures the keyboard-interactive authentication method.
type KeyboardInteractiveAuthConfig struct {
	// Method is the authenticator to use for public keys.
	Method KeyboardInteractiveAuthMethod `json:"method" yaml:"method" default:""`

	// Webhook configures the oAuth2 authenticator for keyboard-interactive authentication.
	OAuth2 AuthOAuth2ClientConfig `json:"oauth2" yaml:"oauth2"`
}

func (c KeyboardInteractiveAuthConfig) Validate() error {
	if err := c.Method.Validate(); err != nil {
		return wrap(err, "method")
	}
	switch c.Method {
	case KeyboardInteractiveAuthMethodDisabled:
		return nil
	case KeyboardInteractiveAuthMethodOAuth2:
		return wrap(c.OAuth2.Validate(), "oauth2")
	default:
		return newError("method", "BUG: unsupported keyboard-interactive authentication method: %s", c.Method)
	}
}

// KeyboardInteractiveAuthMethod provides the methods usable for keyboard-interactive authentication.
type KeyboardInteractiveAuthMethod string

// Validate checks if the provided method is valid or not.
func (m KeyboardInteractiveAuthMethod) Validate() error {
	if m == KeyboardInteractiveAuthMethodDisabled || m == KeyboardInteractiveAuthMethodOAuth2 {
		return nil
	}
	return fmt.Errorf("invalid value for method for keyboard-interactive authentication: %s", m)
}

// KeyboardInteractiveAuthMethodDisabled disables keyboard-interactive authentication.
const KeyboardInteractiveAuthMethodDisabled KeyboardInteractiveAuthMethod = KeyboardInteractiveAuthMethod(AuthMethodDisabled)

// KeyboardInteractiveAuthMethodOAuth2 authenticates using oAuth2/OIDC.
const KeyboardInteractiveAuthMethodOAuth2 KeyboardInteractiveAuthMethod = KeyboardInteractiveAuthMethod(AuthMethodOAuth2)

// endregion

// region GSSAPI

type GSSAPIAuthConfig struct {
	// Method is the authenticator to use for GSSAPI authentication.
	Method GSSAPIAuthMethod `json:"method" yaml:"method"`

	// Kerberos configures GSSAPI for Kerberos authentication.
	Kerberos AuthKerberosClientConfig `json:"kerberos" yaml:"kerberos"`
}

func (c GSSAPIAuthConfig) Validate() error {
	if err := c.Method.Validate(); err != nil {
		return wrap(err, "method")
	}
	switch c.Method {
	case GSSAPIAuthMethodDisabled:
		return nil
	case GSSAPIAuthMethodKerberos:
		return wrap(c.Kerberos.Validate(), "kerberos")
	default:
		return newError("method", "BUG: unsupported GSSAPI authentication method: %s", c.Method)
	}
}

// GSSAPIAuthMethod provides the methods usable for GSSAPI authentication.
type GSSAPIAuthMethod string

// Validate checks if the provided method is valid or not.
func (m GSSAPIAuthMethod) Validate() error {
	if m == GSSAPIAuthMethodDisabled || m == GSSAPIAuthMethodKerberos {
		return nil
	}
	return fmt.Errorf("invalid value for method for GSSAPI authentication: %s", m)
}

// GSSAPIAuthMethodDisabled disables GSSAPI authentication.
const GSSAPIAuthMethodDisabled GSSAPIAuthMethod = GSSAPIAuthMethod(AuthMethodDisabled)

// GSSAPIAuthMethodKerberos authenticates using Kerberos.
const GSSAPIAuthMethodKerberos GSSAPIAuthMethod = GSSAPIAuthMethod(AuthMethodKerberos)

// endregion

// region Webhook

// AuthWebhookClientConfig is the configuration for webhook authentication.
type AuthWebhookClientConfig struct {
	HTTPClientConfiguration `json:",inline" yaml:",inline"`

	// AuthTimeout is the timeout for the overall authentication call (e.g. verifying a password). If the server
	// responds with a non-200 response the call will be retried until this timeout is reached.
	AuthTimeout time.Duration `json:"authTimeout" yaml:"authTimeout" default:"60s"`
}

// Validate validates the authentication client configuration.
func (c *AuthWebhookClientConfig) Validate() error {
	if c.AuthTimeout < 100*time.Millisecond {
		return newError("timeout", "auth timeout value %s is too low, must be at least 100ms", c.AuthTimeout.String())
	}
	if err := c.HTTPClientConfiguration.Validate(); err != nil {
		return err
	}
	return nil
}

// endregion

// region oAuth2

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

	// Generic is a generic OAuth2 authentication without OIDC userinfo. This method cannot verify the username.
	Generic AuthGenericConfig `json:"generic" yaml:"generic"`

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

	// AuthTimeout is the timeout for the overall authentication call. If the server
	// responds with a non-200 response the call will be retried until this timeout is reached.
	AuthTimeout time.Duration `json:"authTimeout" yaml:"authTimeout" default:"120s"`
}

// Validate validates if the OAuth2 client configuration is valid.
func (o *AuthOAuth2ClientConfig) Validate() error {
	if err := o.Redirect.Validate(); err != nil {
		return wrap(err, "redirect")
	}
	if o.ClientID == "" {
		return newError("clientId", "empty client ID")
	}

	if o.ClientSecret == "" {
		return newError("clientSecret", "empty client secret")
	}

	if err := o.Provider.Validate(); err != nil {
		return wrap(err, "provider")
	}

	switch o.Provider {
	case AuthOAuth2GitHubProvider:
		if err := o.GitHub.Validate(); err != nil {
			return wrap(err, "github")
		}
	case AuthOAuth2OIDCProvider:
		if err := o.OIDC.Validate(); err != nil {
			return wrap(err, "oidc")
		}
	case AuthOAuth2GenericProvider:
		if err := o.Generic.Validate(); err != nil {
			return wrap(err, "generic")
		}
	}

	return nil
}

// OAuth2ProviderName provides the various methods of oAuth2 authentication.
type OAuth2ProviderName string

const (
	// AuthOAuth2GitHubProvider authenticates against the GitHub API or a private GitHub Enterprise instance.
	AuthOAuth2GitHubProvider OAuth2ProviderName = "github"
	// AuthOAuth2OIDCProvider authenticates against a configured OIDC server.
	AuthOAuth2OIDCProvider OAuth2ProviderName = "oidc"
	// AuthOAuth2GenericProvider authenticates against a generic OAuth2 server.
	AuthOAuth2GenericProvider OAuth2ProviderName = "generic"
)

func (o OAuth2ProviderName) Validate() error {
	switch o {
	case AuthOAuth2GitHubProvider:
		return nil
	case AuthOAuth2OIDCProvider:
		return nil
	case AuthOAuth2GenericProvider:
		return nil
	default:
		return fmt.Errorf("invalid oAuth2 provider")
	}
}

// OAuth2RedirectConfig is the configuration for the HTTP server that serves the page presented to the user after they
// are authenticated.
type OAuth2RedirectConfig struct {
	HTTPServerConfiguration `json:",inline" yaml:",inline"`

	// Webroot is a directory which contains all files that should be served as part of the return page
	// the user lands on when completing the oAuth2 authentication process. The webroot must contain an
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
			return wrap(err, "webroot")
		}
		if !webrootStat.IsDir() {
			return newError("webroot", "invalid webroot (not a directory)")
		}
		indexStat, err := os.Stat(path.Join(o.Webroot, "index.html"))
		if err != nil {
			return wrapWithMessage(err, "webroot", "webroot does not contain an index.html file")
		}
		if indexStat.IsDir() {
			return newError("webroot", "webroot does not contain an index.html file (index.html is a directory)")
		}
	}
	return nil
}

// AuthGitHubConfig is the configuration structure for GitHub authentication.
type AuthGitHubConfig struct {
	// CACert is the PEM-encoded CA certificate, or file containing a PEM-encoded CA certificate used to verify the
	// GitHub server certificate.
	CACert string `json:"cacert" yaml:"cacert" default:""`

	// URL is the base GitHub URL. Change this for GitHub Enterprise.
	URL string `json:"url" yaml:"url" default:"https://github.com"`

	// APIURL is the GitHub API URL. Change this for GitHub Enterprise.
	APIURL string `json:"apiurl" yaml:"apiurl" default:"https://api.github.com"`

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
	// the configuration server has to handle the GITHUB_USER connnection parameter in order to obtain the correct
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
		return wrap(err, "url")
	}

	return nil
}

type AuthOIDCConfig struct {
	HTTPClientConfiguration `json:",inline" yaml:",inline"`

	// DeviceFlow enables or disables using the OIDC device flow.
	DeviceFlow bool `json:"deviceFlow" yaml:"deviceFlow"`
	// AuthorizationCodeFlow enables or disables the OIDC authorization code flow.
	AuthorizationCodeFlow bool `json:"authorizationCodeFlow" yaml:"authorizationCodeFlow"`

	// UsernameField indicates the userinfo field that should be taken as the username.
	UsernameField string `json:"usernameField" yaml:"usernameField" default:"sub"`

	// RedirectURI is the URI the client is returned to. This URL should be configured to the redirect server endpoint.
	RedirectURI string `json:"redirectURI" yaml:"redirectURI"`
}

func (o *AuthOIDCConfig) Validate() error {
	if !o.DeviceFlow && !o.AuthorizationCodeFlow {
		return fmt.Errorf("at least one of deviceFlow or authorizationCodeFlow must be enabled")
	}
	if o.AuthorizationCodeFlow && o.RedirectURI == "" {
		return wrap(fmt.Errorf("redirectURI is required if the authorization code flow is enabled"), "redirectURI")
	}
	return o.HTTPClientConfiguration.Validate()
}

type AuthGenericConfig struct {
	// AuthorizeEndpointURL is the endpoint configuration for the OAuth2 authorization
	AuthorizeEndpointURL string `json:"authorizeEndpointURL" yaml:"authorizeEndpointURL"`
	// TokenEndpoint is the endpoint configuration for the OAuth2 authorization
	TokenEndpoint HTTPClientConfiguration `json:"tokenEndpoint" yaml:"tokenEndpoint"`
	// RedirectURI is the URI the client is returned to. This URL should be configured to the redirect server endpoint.
	RedirectURI string `json:"redirectURI" yaml:"redirectURI"`
}

func (o *AuthGenericConfig) Validate() error {
	if _, err := url.ParseRequestURI(o.AuthorizeEndpointURL); err != nil {
		return newError("authorizeEndpointURL", "invalid URL: %s", o.AuthorizeEndpointURL)
	}

	if _, err := url.ParseRequestURI(o.RedirectURI); err != nil {
		return newError("redirectURI", "invalid URL: %s", o.RedirectURI)
	}

	if err := o.TokenEndpoint.Validate(); err != nil {
		return wrap(err, "tokenEndpoint")
	}
	return nil
}

// endregion

// region Authz

// AuthzConfig is the configuration for the authorization flow
type AuthzConfig struct {
	Method AuthzMethod `json:"method" yaml:"method" default:""`

	AuthWebhookClientConfig `json:",inline" yaml:",inline"`
}

// Validate validates the authorization configuration.
func (k *AuthzConfig) Validate() error {
	if err := k.Method.Validate(); err != nil {
		return wrap(err, "method")
	}
	switch k.Method {
	case AuthzMethodDisabled:
		return nil
	case AuthzMethodWebhook:
		return wrap(k.AuthWebhookClientConfig.Validate(), "webhook")
	default:
		return newError("method", "BUG: invalid value for method for authorization: %s", k.Method)
	}
}

// AuthzMethod provides the methods usable authorization.
type AuthzMethod string

// Validate checks if the provided method is valid or not.
func (m AuthzMethod) Validate() error {
	if m == AuthzMethodDisabled || m == AuthzMethodWebhook {
		return nil
	}
	return fmt.Errorf("invalid value for method for authorization: %s", m)
}

// AuthzMethodDisabled disables authorization. All usernames will be accepted as they are authenticated.
const AuthzMethodDisabled AuthzMethod = AuthzMethod(AuthMethodDisabled)

// AuthzMethodWebhook authorizes users using HTTP webhooks.
const AuthzMethodWebhook AuthzMethod = AuthzMethod(AuthMethodWebhook)

// endregion

// region Kerberos

// AuthKerberosClientConfig is the configuration for the Kerberos authentication method.
type AuthKerberosClientConfig struct {
	// Keytab is the path to the kerberos keytab. If unset it defaults to
	// the default of /etc/krb5.keytab. If this file doesn't exist and
	// kerberos authentication is requested ContainerSSH will fail to start.
	Keytab string `json:"keytab" yaml:"keytab" default:"/etc/krb5.keytab"`
	// Acceptor is the name of the keytab entry to authenticate against.
	// The value of this field needs to be in the form of `service/name`.
	//
	// The special value of `host` will authenticate clients only against
	// the service `host/hostname` where hostname is the system hostname
	// The special value of 'any' will authenticate against all keytab
	// entries regardless of name.
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
	// ConfigPath is the path of the kerberos configuration file. This is
	// only used for password authentication.
	ConfigPath string `json:"configPath" yaml:"configPath" default:"/etc/containerssh/krb5.conf"`
	// ClockSkew is the maximum allowed clock skew for kerberos messages,
	// any messages older than this will be rejected. This value is also
	// used for the replay cache.
	ClockSkew time.Duration `json:"clockSkew" yaml:"clockSkew" default:"5m"`
}

func (k *AuthKerberosClientConfig) Validate() error {
	if _, err := os.Stat(k.Keytab); err != nil {
		return wrapWithMessage(err, "keytab file %s does not exist or is inaccessible", k.Keytab)
	}
	return nil
}

// endregion
