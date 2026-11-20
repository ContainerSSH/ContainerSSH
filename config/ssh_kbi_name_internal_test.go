package config

import "testing"

func TestValidateKeyboardInteractiveNameEmptyIsValid(t *testing.T) {
	if err := validateKeyboardInteractiveName(""); err != nil {
		t.Fatalf("expected an empty keyboardInteractiveName to be valid, got: %v", err)
	}
}

func TestValidateKeyboardInteractiveNameTemplates(t *testing.T) {
	if err := validateKeyboardInteractiveName("{{.Username}}"); err != nil {
		t.Fatalf("expected the username template to be valid, got: %v", err)
	}
	if err := validateKeyboardInteractiveName("Example Corp SSH login"); err != nil {
		t.Fatalf("expected a static name to be valid, got: %v", err)
	}
	if err := validateKeyboardInteractiveName("{{.Username"); err == nil {
		t.Fatal("expected an error for an unterminated template, got none")
	}
	if err := validateKeyboardInteractiveName("{{.NoSuchField}}"); err == nil {
		t.Fatal("expected an error for an unknown template variable, got none")
	}
}

func newTestSSHConfigForKbiName() SSHConfig {
	return SSHConfig{
		ServerVersion:       "SSH-2.0-ContainerSSH",
		Ciphers:             SSHCipherList{"aes256-ctr"},
		KexAlgorithms:       SSHKexList{"curve25519-sha256@libssh.org"},
		MACs:                SSHMACList{"hmac-sha2-256"},
		ClientAliveCountMax: 3,
	}
}

func TestSSHConfigValidateKeyboardInteractiveName(t *testing.T) {
	cfg := newTestSSHConfigForKbiName()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected an unset keyboardInteractiveName to be valid, got: %v", err)
	}

	name := "{{.Username}}"
	cfg.KeyboardInteractiveName = &name
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected the username template to be valid, got: %v", err)
	}

	broken := "{{.Username"
	cfg.KeyboardInteractiveName = &broken
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected an error for an unterminated template, got none")
	}
}
