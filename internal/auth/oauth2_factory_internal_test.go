package auth

import (
	"strings"
	"testing"

	"go.containerssh.io/containerssh/config"
	"go.containerssh.io/containerssh/internal/structutils"
	"go.containerssh.io/containerssh/log"
)

func newTestOAuth2ClientConfig() config.AuthOAuth2ClientConfig {
	cfg := config.AuthOAuth2ClientConfig{}
	structutils.Defaults(&cfg)
	cfg.ClientID = "test-client"
	cfg.ClientSecret = "test-secret"
	cfg.Provider = config.AuthOAuth2GitHubProvider
	return cfg
}

func TestNewOAuth2ClientDefaultPrompts(t *testing.T) {
	cfg := newTestOAuth2ClientConfig()

	client, redirectServer, err := NewOAuth2Client(cfg, log.NewTestLogger(t), nil)
	if err != nil {
		t.Fatalf("failed to create the oAuth2 client (%v)", err)
	}
	if redirectServer == nil {
		t.Fatal("expected a redirect server")
	}
	oauth2client, ok := client.(*oauth2Client)
	if !ok {
		t.Fatalf("unexpected client type: %T", client)
	}
	prompt, err := renderOAuth2Prompt(oauth2client.deviceFlowPrompt, config.OAuth2DeviceFlowPromptData{
		AuthorizationURL: "https://example.com/device",
		UserCode:         "ABCD-EFGH",
	})
	if err != nil {
		t.Fatalf("failed to render the default device flow prompt (%v)", err)
	}
	if !strings.HasPrefix(prompt, "Please click the following link: https://example.com/device") {
		t.Fatalf("unexpected default device flow prompt: %q", prompt)
	}
}

func TestNewOAuth2ClientCustomPrompts(t *testing.T) {
	cfg := newTestOAuth2ClientConfig()
	cfg.DeviceFlowPrompt = "Open {{.AuthorizationURL}} and confirm {{.UserCode}}."
	cfg.AuthorizationCodeFlowPrompt = "Log in: {{.AuthorizationURL}}"

	client, _, err := NewOAuth2Client(cfg, log.NewTestLogger(t), nil)
	if err != nil {
		t.Fatalf("failed to create the oAuth2 client (%v)", err)
	}
	oauth2client := client.(*oauth2Client)
	prompt, err := renderOAuth2Prompt(oauth2client.deviceFlowPrompt, config.OAuth2DeviceFlowPromptData{
		AuthorizationURL: "https://example.com/device",
		UserCode:         "ABCD-EFGH",
	})
	if err != nil {
		t.Fatalf("failed to render the custom device flow prompt (%v)", err)
	}
	if prompt != "Open https://example.com/device and confirm ABCD-EFGH." {
		t.Fatalf("unexpected custom device flow prompt: %q", prompt)
	}
	prompt, err = renderOAuth2Prompt(oauth2client.authorizationCodeFlowPrompt, config.OAuth2AuthorizationCodeFlowPromptData{
		AuthorizationURL: "https://example.com/login",
	})
	if err != nil {
		t.Fatalf("failed to render the custom authorization code flow prompt (%v)", err)
	}
	if prompt != "Log in: https://example.com/login" {
		t.Fatalf("unexpected custom authorization code flow prompt: %q", prompt)
	}
}

func TestNewOAuth2ClientRejectsInvalidPrompt(t *testing.T) {
	cfg := newTestOAuth2ClientConfig()
	cfg.DeviceFlowPrompt = "{{.UserCode"

	if _, _, err := NewOAuth2Client(cfg, log.NewTestLogger(t), nil); err == nil {
		t.Fatal("expected an error for an invalid device flow prompt, got none")
	}
}
