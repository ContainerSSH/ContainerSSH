package message

import "bytes"

// PayloadAuthPassword is a payload for a message that indicates an authentication attempt, successful, or failed
// authentication.
type PayloadAuthPassword struct {
	Username string `json:"username" yaml:"username"`
	Password []byte `json:"password" yaml:"password"`
}

// Equals compares two PayloadAuthPassword payloads.
func (p PayloadAuthPassword) Equals(other Payload) bool {
	p2, ok := other.(PayloadAuthPassword)
	if !ok {
		return false
	}
	return p.Username == p2.Username && bytes.Equal(p.Password, p2.Password)
}

// PayloadAuthPasswordSuccessful is a payload for a message that indicates a successful authentication with a
// password.
type PayloadAuthPasswordSuccessful struct {
	Username              string                         `json:"username" yaml:"username"`
	AuthenticatedUsername string                         `json:"authenticatedUsername" yaml:"authenticatedUsername"`
	Password              []byte                         `json:"password" yaml:"password"`
	Metadata              map[string]MetadataValue       `json:"metadata,omitempty"`
	Environment           map[string]MetadataValue       `json:"environment,omitempty"`
	Files                 map[string]MetadataBinaryValue `json:"files,omitempty"`
}

// Equals compares two PayloadAuthPasswordSuccessful payloads.
func (p PayloadAuthPasswordSuccessful) Equals(other Payload) bool {
	p2, ok := other.(PayloadAuthPasswordSuccessful)
	if !ok {
		return false
	}
	return p.Username == p2.Username &&
		p.AuthenticatedUsername == p2.AuthenticatedUsername &&
		bytes.Equal(
			p.Password,
			p2.Password,
		)
}

// PayloadAuthPasswordBackendError is a payload for a message that indicates a backend failure during authentication.
type PayloadAuthPasswordBackendError struct {
	Username string `json:"username" yaml:"username"`
	Password []byte `json:"password" yaml:"password"`
	Reason   string `json:"reason" yaml:"reason"`
}

// Equals compares two PayloadAuthPasswordBackendError payloads.
func (p PayloadAuthPasswordBackendError) Equals(other Payload) bool {
	p2, ok := other.(PayloadAuthPasswordBackendError)
	if !ok {
		return false
	}
	return p.Username == p2.Username && bytes.Equal(p.Password, p2.Password) && p.Reason == p2.Reason
}

// PayloadAuthPubKey is a payload for a public key based authentication.
type PayloadAuthPubKey struct {
	Username string `json:"username" yaml:"username"`
	Key      string `json:"key" yaml:"key"`
}

// Equals compares two PayloadAuthPubKey payloads.
func (p PayloadAuthPubKey) Equals(other Payload) bool {
	p2, ok := other.(PayloadAuthPubKey)
	if !ok {
		return false
	}
	return p.Username == p2.Username && p.Key == p2.Key
}

// PayloadAuthPubKeyBackendError is a payload for a message indicating that there was a backend error while
// authenticating with public key.
type PayloadAuthPubKeyBackendError struct {
	Username string `json:"username" yaml:"username"`
	Key      string `json:"key" yaml:"key"`
	Reason   string `json:"reason" yaml:"reason"`
}

// Equals compares two PayloadAuthPubKeyBackendError payloads.
func (p PayloadAuthPubKeyBackendError) Equals(other Payload) bool {
	p2, ok := other.(PayloadAuthPubKeyBackendError)
	if !ok {
		return false
	}
	return p.Username == p2.Username && p.Key == p2.Key && p.Reason == p2.Reason
}

// PayloadAuthKeyboardInteractiveChallenge is a message that indicates that a keyboard-interactive challenge has been
// sent to the user. Multiple challenge-response interactions can take place.
type PayloadAuthKeyboardInteractiveChallenge struct {
	Username    string                        `json:"username" yaml:"username"`
	Instruction string                        `json:"instruction" yaml:"instruction"`
	Questions   []KeyboardInteractiveQuestion `json:"questions" yaml:"questions"`
}

// Equals compares two PayloadAuthKeyboardInteractiveChallenge messages.
func (p PayloadAuthKeyboardInteractiveChallenge) Equals(other Payload) bool {
	p2, ok := other.(PayloadAuthKeyboardInteractiveChallenge)
	if !ok {
		return false
	}
	if p.Username != p2.Username {
		return false
	}
	if p.Instruction != p2.Instruction {
		return false
	}
	if len(p.Questions) != len(p2.Questions) {
		return false
	}
	for i, question := range p.Questions {
		if !question.Equals(p2.Questions[i]) {
			return false
		}
	}
	return true
}

// PayloadAuthKeyboardInteractiveAnswer is a message that indicates a response to a keyboard-interactive challenge.
type PayloadAuthKeyboardInteractiveAnswer struct {
	Username string                      `json:"username" yaml:"username"`
	Answers  []KeyboardInteractiveAnswer `json:"answers" yaml:"answers"`
}

// Equals compares two PayloadAuthKeyboardInteractiveAnswer messages.
func (p PayloadAuthKeyboardInteractiveAnswer) Equals(other Payload) bool {
	p2, ok := other.(PayloadAuthKeyboardInteractiveAnswer)
	if !ok {
		return false
	}
	if p.Username != p2.Username {
		return false
	}
	if len(p.Answers) != len(p2.Answers) {
		return false
	}
	for i, answer := range p.Answers {
		if !answer.Equals(p2.Answers[i]) {
			return false
		}
	}
	return true
}

// PayloadAuthKeyboardInteractiveFailed indicates that a keyboard-interactive authentication process has failed.
type PayloadAuthKeyboardInteractiveFailed struct {
	Username string
}

// Equals compares two PayloadAuthKeyboardInteractiveFailed payloads.
func (p PayloadAuthKeyboardInteractiveFailed) Equals(other Payload) bool {
	p2, ok := other.(PayloadAuthKeyboardInteractiveFailed)
	if !ok {
		return false
	}
	return p.Username == p2.Username
}

// PayloadAuthKeyboardInteractiveBackendError indicates an error in the authentication backend during a
// keyboard-interactive authentication.
type PayloadAuthKeyboardInteractiveBackendError struct {
	Username string `json:"username" yaml:"username"`
	Reason   string `json:"reason" yaml:"reason"`
}

// Equals compares two PayloadAuthKeyboardInteractiveBackendError payloads.
func (p PayloadAuthKeyboardInteractiveBackendError) Equals(other Payload) bool {
	p2, ok := other.(PayloadAuthKeyboardInteractiveBackendError)
	if !ok {
		return false
	}
	return p.Username == p2.Username && p.Reason == p2.Reason
}

// PayloadHandshakeFailed is a payload for a failed handshake.
type PayloadHandshakeFailed struct {
	Reason string `json:"reason" yaml:"reason"`
}

// Equals compares two PayloadHandshakeFailed payloads.
func (p PayloadHandshakeFailed) Equals(other Payload) bool {
	p2, ok := other.(PayloadHandshakeFailed)
	if !ok {
		return false
	}
	return p.Reason == p2.Reason
}

// PayloadHandshakeSuccessful is a payload for a successful handshake.
type PayloadHandshakeSuccessful struct {
	Username string `json:"username" yaml:"username"`
}

// Equals compares two PayloadHandshakeSuccessful payloads.
func (p PayloadHandshakeSuccessful) Equals(other Payload) bool {
	p2, ok := other.(PayloadHandshakeSuccessful)
	if !ok {
		return false
	}
	return p.Username == p2.Username
}

// KeyboardInteractiveQuestion is a description of a question during a keyboard-interactive authentication.
type KeyboardInteractiveQuestion struct {
	Question string `json:"question" yaml:"question"` // The question text sent to the user.
	Echo     bool   `json:"echo" yaml:"echo"`         // True if the input was visible on the screen.
}

// Equals compares two KeyboardInteractiveQuestion submessages.
func (q KeyboardInteractiveQuestion) Equals(q2 KeyboardInteractiveQuestion) bool {
	return q.Question == q2.Question && q.Echo == q2.Echo
}

// KeyboardInteractiveAnswer is the response from the user to a keyboard-interactive authentication.
type KeyboardInteractiveAnswer struct {
	Question string `json:"question" yaml:"question"` // The question text sent to the user.
	Answer   string `json:"answer" yaml:"question"`   // The response from the user.
}

// Equals compares two KeyboardInteractiveAnswer submessages.
func (k KeyboardInteractiveAnswer) Equals(k2 KeyboardInteractiveAnswer) bool {
	return k.Question == k2.Question && k.Answer == k2.Answer
}
