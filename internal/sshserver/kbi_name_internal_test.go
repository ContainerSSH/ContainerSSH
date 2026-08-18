package sshserver

import (
	"testing"
	"text/template"

	"go.containerssh.io/containerssh/config"
	"go.containerssh.io/containerssh/internal/structutils"
	"go.containerssh.io/containerssh/log"
)

func TestRenderKeyboardInteractiveNameUnset(t *testing.T) {
	server := &serverImpl{}
	name, err := server.renderKeyboardInteractiveName("foo")
	if err != nil {
		t.Fatalf("expected no error for an unset template, got: %v", err)
	}
	if name != "" {
		t.Fatalf("expected an empty name for an unset template, got: %q", name)
	}
}

func TestRenderKeyboardInteractiveNameUsername(t *testing.T) {
	server := &serverImpl{
		kbiNameTemplate: template.Must(template.New("keyboardInteractiveName").Parse("{{.Username}}")),
	}
	name, err := server.renderKeyboardInteractiveName("foo")
	if err != nil {
		t.Fatalf("expected no error for the username template, got: %v", err)
	}
	if name != "foo" {
		t.Fatalf("expected the username as name, got: %q", name)
	}
}

func TestRenderKeyboardInteractiveNameStatic(t *testing.T) {
	server := &serverImpl{
		kbiNameTemplate: template.Must(template.New("keyboardInteractiveName").Parse("Example Corp SSH login")),
	}
	name, err := server.renderKeyboardInteractiveName("foo")
	if err != nil {
		t.Fatalf("expected no error for a static name, got: %v", err)
	}
	if name != "Example Corp SSH login" {
		t.Fatalf("expected the static name, got: %q", name)
	}
}

func TestNewParsesKeyboardInteractiveName(t *testing.T) {
	cfg := config.SSHConfig{}
	structutils.Defaults(&cfg)
	if err := cfg.GenerateHostKey(); err != nil {
		t.Fatalf("failed to generate a host key (%v)", err)
	}

	// Default: the previous behavior, sending the username.
	srv, err := New(cfg, nil, log.NewTestLogger(t))
	if err != nil {
		t.Fatalf("failed to create the SSH server (%v)", err)
	}
	name, err := srv.(*serverImpl).renderKeyboardInteractiveName("foo")
	if err != nil {
		t.Fatalf("failed to render the default keyboard-interactive name (%v)", err)
	}
	if name != "foo" {
		t.Fatalf("expected the username as the default name, got: %q", name)
	}

	// Empty string: no name is sent.
	empty := ""
	cfg.KeyboardInteractiveName = &empty
	srv, err = New(cfg, nil, log.NewTestLogger(t))
	if err != nil {
		t.Fatalf("failed to create the SSH server (%v)", err)
	}
	if srv.(*serverImpl).kbiNameTemplate != nil {
		t.Fatal("expected no name template for an empty keyboardInteractiveName")
	}

	// Invalid template: rejected by the configuration validation.
	broken := "{{.Username"
	cfg.KeyboardInteractiveName = &broken
	if _, err := New(cfg, nil, log.NewTestLogger(t)); err == nil {
		t.Fatal("expected an error for an unterminated keyboardInteractiveName template, got none")
	}
}
