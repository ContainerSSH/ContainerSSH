package config

import "testing"

func TestValidateOAuth2PromptEmptyIsValid(t *testing.T) {
	if err := validateOAuth2Prompt("", OAuth2DeviceFlowPromptData{}); err != nil {
		t.Fatalf("expected an empty prompt to be valid, got: %v", err)
	}
}

func TestValidateOAuth2PromptDefaultsAreValid(t *testing.T) {
	if err := validateOAuth2Prompt(DefaultDeviceFlowPrompt, OAuth2DeviceFlowPromptData{}); err != nil {
		t.Fatalf("expected the default device flow prompt to be valid, got: %v", err)
	}
	if err := validateOAuth2Prompt(
		DefaultAuthorizationCodeFlowPrompt,
		OAuth2AuthorizationCodeFlowPromptData{},
	); err != nil {
		t.Fatalf("expected the default authorization code flow prompt to be valid, got: %v", err)
	}
}

func TestValidateOAuth2PromptRejectsBrokenTemplates(t *testing.T) {
	if err := validateOAuth2Prompt("{{.UserCode", OAuth2DeviceFlowPromptData{}); err == nil {
		t.Fatal("expected an error for an unterminated template, got none")
	}
	if err := validateOAuth2Prompt("{{.NoSuchField}}", OAuth2DeviceFlowPromptData{}); err == nil {
		t.Fatal("expected an error for an unknown template variable, got none")
	}
}

func newTestAuthOAuth2ClientConfig() AuthOAuth2ClientConfig {
	return AuthOAuth2ClientConfig{
		Redirect: OAuth2RedirectConfig{
			HTTPServerConfiguration: HTTPServerConfiguration{
				Listen: "127.0.0.1:8080",
			},
		},
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		Provider:     AuthOAuth2GitHubProvider,
		GitHub: AuthGitHubConfig{
			URL: "https://github.com",
		},
	}
}

func TestAuthOAuth2ClientConfigValidatePrompts(t *testing.T) {
	cfg := newTestAuthOAuth2ClientConfig()
	cfg.DeviceFlowPrompt = "Open {{.AuthorizationURL}} and confirm {{.UserCode}}."
	cfg.AuthorizationCodeFlowPrompt = "Log in: {{.AuthorizationURL}}"
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid prompts to pass validation, got: %v", err)
	}

	cfg = newTestAuthOAuth2ClientConfig()
	cfg.DeviceFlowPrompt = "{{.UserCode"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected an error for an invalid deviceFlowPrompt, got none")
	}

	cfg = newTestAuthOAuth2ClientConfig()
	cfg.AuthorizationCodeFlowPrompt = "{{.NoSuchField}}"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected an error for an invalid authorizationCodeFlowPrompt, got none")
	}
}
