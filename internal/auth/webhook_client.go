// This file contains the details of the webhook authenticator

package auth

// WebhookClient is a urlEncodedClient that authenticates using HTTP webhooks. It only supports password and public key
// authentication.
type WebhookClient interface {
	PasswordAuthenticator
	PublicKeyAuthenticator
	AuthzProvider
}
