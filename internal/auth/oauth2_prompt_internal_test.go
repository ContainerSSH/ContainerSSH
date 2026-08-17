package auth

import (
	"testing"

	"go.containerssh.io/containerssh/config"
)

func TestRenderOAuth2PromptDefaultDeviceFlow(t *testing.T) {
	tpl, err := parseOAuth2Prompt("deviceFlowPrompt", "", config.DefaultDeviceFlowPrompt)
	if err != nil {
		t.Fatalf("failed to parse the default device flow prompt (%v)", err)
	}
	prompt, err := renderOAuth2Prompt(tpl, config.OAuth2DeviceFlowPromptData{
		AuthorizationURL: "https://example.com/device",
		UserCode:         "ABCD-EFGH",
	})
	if err != nil {
		t.Fatalf("failed to render the default device flow prompt (%v)", err)
	}
	expected := "Please click the following link: https://example.com/device\n\nEnter the following code: ABCD-EFGH\n"
	if prompt != expected {
		t.Fatalf("unexpected device flow prompt: %q", prompt)
	}
}

func TestRenderOAuth2PromptDefaultAuthorizationCodeFlow(t *testing.T) {
	tpl, err := parseOAuth2Prompt("authorizationCodeFlowPrompt", "", config.DefaultAuthorizationCodeFlowPrompt)
	if err != nil {
		t.Fatalf("failed to parse the default authorization code flow prompt (%v)", err)
	}
	prompt, err := renderOAuth2Prompt(tpl, config.OAuth2AuthorizationCodeFlowPromptData{
		AuthorizationURL: "https://example.com/login",
	})
	if err != nil {
		t.Fatalf("failed to render the default authorization code flow prompt (%v)", err)
	}
	expected := "Please click the following link to log in: https://example.com/login\n\n"
	if prompt != expected {
		t.Fatalf("unexpected authorization code flow prompt: %q", prompt)
	}
}

func TestRenderOAuth2PromptCustomTemplate(t *testing.T) {
	tpl, err := parseOAuth2Prompt(
		"deviceFlowPrompt",
		"Welcome to example.org!\n\nOpen {{.AuthorizationURL}} and confirm code {{.UserCode}}.\n",
		config.DefaultDeviceFlowPrompt,
	)
	if err != nil {
		t.Fatalf("failed to parse the custom device flow prompt (%v)", err)
	}
	prompt, err := renderOAuth2Prompt(tpl, config.OAuth2DeviceFlowPromptData{
		AuthorizationURL: "https://example.org/device",
		UserCode:         "WXYZ-1234",
	})
	if err != nil {
		t.Fatalf("failed to render the custom device flow prompt (%v)", err)
	}
	expected := "Welcome to example.org!\n\nOpen https://example.org/device and confirm code WXYZ-1234.\n"
	if prompt != expected {
		t.Fatalf("unexpected custom device flow prompt: %q", prompt)
	}
}

func TestParseOAuth2PromptInvalidTemplate(t *testing.T) {
	if _, err := parseOAuth2Prompt("deviceFlowPrompt", "{{.UserCode", config.DefaultDeviceFlowPrompt); err == nil {
		t.Fatal("expected an error for an unterminated template, got none")
	}
}
