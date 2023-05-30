package sshserver

import (
	"bytes"
	cryptoRand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"go.containerssh.io/libcontainerssh/internal/test"
	"golang.org/x/crypto/ssh"
)

// TestUser is a container for a username, a password and public keys
type TestUser struct {
	username            string
	password            string
	keyboardInteractive map[string]string
	privateKeys         []*rsa.PrivateKey
	authorizedKeys      []string
}

// Username returns the username of this user.
func (u *TestUser) Username() string {
	return u.username
}

// Password returns the current password for this user.
func (u *TestUser) Password() string {
	return u.password
}

// SetPassword sets a specific password for this user.
func (u *TestUser) SetPassword(password string) {
	u.password = password
}

// AddKeyboardInteractiveChallengeResponse adds a challenge with an expected response for keyboard-interactive
// authentication.
func (u *TestUser) AddKeyboardInteractiveChallengeResponse(challenge string, expectedResponse string) {
	u.keyboardInteractive[challenge] = expectedResponse
}

// KeyboardInteractiveChallengeResponse returns a construct of KeyboardInteractiveQuestions
func (u *TestUser) KeyboardInteractiveChallengeResponse() (questions KeyboardInteractiveQuestions) {
	for question := range u.keyboardInteractive {
		questions = append(questions, KeyboardInteractiveQuestion{
			Question:     question,
			EchoResponse: false,
		})
	}
	return
}

// RandomPassword generates a random password for this user.
func (u *TestUser) RandomPassword() {
	u.password = test.RandomString(16)
}

// GenerateKey generates a public and private key pair that can be used to authenticate with this user.
func (u *TestUser) GenerateKey() (privateKeyPEM string, publicKeyAuthorizedKeys string) {
	privateKey, err := rsa.GenerateKey(cryptoRand.Reader, 4096)
	if err != nil {
		panic(err)
	}

	privPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	var privateKeyPEMBuffer bytes.Buffer
	if err := pem.Encode(&privateKeyPEMBuffer, privPEM); err != nil {
		panic(err)
	}
	privateKeyPEM = privateKeyPEMBuffer.String()

	pub, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		panic(err)
	}
	publicKeyAuthorizedKeys = string(ssh.MarshalAuthorizedKey(pub))

	u.privateKeys = append(u.privateKeys, privateKey)
	u.authorizedKeys = append(u.authorizedKeys, privateKeyPEM)

	return privateKeyPEM, publicKeyAuthorizedKeys
}

// GetAuthorizedKeys returns a slice of the authorized keys of this user.
func (u *TestUser) GetAuthorizedKeys() []string {
	return u.authorizedKeys
}

func (u *TestUser) GetAuthMethods() []ssh.AuthMethod {
	var result []ssh.AuthMethod
	if u.password != "" {
		result = append(result, ssh.Password(u.password))
	}
	var pubKeys []ssh.Signer
	for _, privateKey := range u.privateKeys {
		signer, err := ssh.NewSignerFromKey(privateKey)
		if err != nil {
			panic(err)
		}
		pubKeys = append(pubKeys, signer)
	}
	result = append(result, ssh.PublicKeys(pubKeys...))

	if len(u.keyboardInteractive) > 0 {
		result = append(result, ssh.KeyboardInteractive(
			func(user, instruction string, questions []string, echos []bool) (answers []string, err error) {
				for _, question := range questions {
					if answer, ok := u.keyboardInteractive[question]; ok {
						answers = append(answers, answer)
					} else {
						return nil, fmt.Errorf("unexpected keybord-interactive challenge: %s", question)
					}
				}
				return answers, nil
			}))
	}
	return result
}
