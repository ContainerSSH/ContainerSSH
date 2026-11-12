package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.containerssh.io/containerssh/config"
	"go.containerssh.io/containerssh/log"
	"go.containerssh.io/containerssh/metadata"
)

type fakeDeviceFlow struct {
	meta         metadata.ConnectionAuthPendingMetadata
	deauthorized bool
}

func (f *fakeDeviceFlow) Deauthorize(_ context.Context) {
	f.deauthorized = true
}

func (f *fakeDeviceFlow) GetAuthorizationURL(_ context.Context) (string, string, time.Duration, error) {
	return "https://example.com/device", "ABCD-EFGH", time.Minute, nil
}

func (f *fakeDeviceFlow) Verify(_ context.Context) (string, metadata.ConnectionAuthenticatedMetadata, error) {
	return "access-token", f.meta.Authenticated(f.meta.Username), nil
}

type fakeAuthCodeFlow struct {
	meta          metadata.ConnectionAuthPendingMetadata
	deauthorized  bool
	receivedState string
	receivedCode  string
}

func (f *fakeAuthCodeFlow) Deauthorize(_ context.Context) {
	f.deauthorized = true
}

func (f *fakeAuthCodeFlow) GetAuthorizationURL(_ context.Context) (string, error) {
	return "https://example.com/login", nil
}

func (f *fakeAuthCodeFlow) Verify(_ context.Context, state string, code string) (
	string,
	metadata.ConnectionAuthenticatedMetadata,
	error,
) {
	f.receivedState = state
	f.receivedCode = code
	return "access-token", f.meta.Authenticated(f.meta.Username), nil
}

type fakeOAuth2Provider struct {
	deviceFlow   OAuth2DeviceFlow
	authCodeFlow OAuth2AuthorizationCodeFlow
}

func (f *fakeOAuth2Provider) SupportsDeviceFlow() bool {
	return f.deviceFlow != nil
}

func (f *fakeOAuth2Provider) GetDeviceFlow(_ context.Context, _ metadata.ConnectionAuthPendingMetadata) (
	OAuth2DeviceFlow,
	error,
) {
	return f.deviceFlow, nil
}

func (f *fakeOAuth2Provider) SupportsAuthorizationCodeFlow() bool {
	return f.authCodeFlow != nil
}

func (f *fakeOAuth2Provider) GetAuthorizationCodeFlow(_ context.Context, _ metadata.ConnectionAuthPendingMetadata) (
	OAuth2AuthorizationCodeFlow,
	error,
) {
	return f.authCodeFlow, nil
}

func newTestOAuth2Client(
	t *testing.T,
	provider OAuth2Provider,
	deviceFlowPrompt string,
	authorizationCodeFlowPrompt string,
) *oauth2Client {
	deviceTpl, err := parseOAuth2Prompt("deviceFlowPrompt", deviceFlowPrompt, config.DefaultDeviceFlowPrompt)
	if err != nil {
		t.Fatalf("failed to parse the device flow prompt (%v)", err)
	}
	authCodeTpl, err := parseOAuth2Prompt(
		"authorizationCodeFlowPrompt",
		authorizationCodeFlowPrompt,
		config.DefaultAuthorizationCodeFlowPrompt,
	)
	if err != nil {
		t.Fatalf("failed to parse the authorization code flow prompt (%v)", err)
	}
	return &oauth2Client{
		provider:                    provider,
		logger:                      log.NewTestLogger(t),
		deviceFlowPrompt:            deviceTpl,
		authorizationCodeFlowPrompt: authCodeTpl,
	}
}

func testAuthPendingMetadata() metadata.ConnectionAuthPendingMetadata {
	return metadata.NewTestMetadata().StartAuthentication("SSH-2.0-test", "foo")
}

func TestOAuth2KeyboardInteractiveDeviceFlowDefaultPrompt(t *testing.T) {
	meta := testAuthPendingMetadata()
	flow := &fakeDeviceFlow{meta: meta}
	client := newTestOAuth2Client(t, &fakeOAuth2Provider{deviceFlow: flow}, "", "")

	var instruction string
	ctx := client.KeyboardInteractive(
		meta,
		func(i string, questions KeyboardInteractiveQuestions) (KeyboardInteractiveAnswers, error) {
			instruction = i
			if len(questions) != 0 {
				t.Fatalf("expected no questions for the device flow, got %d", len(questions))
			}
			return KeyboardInteractiveAnswers{}, nil
		},
	)
	if !ctx.Success() {
		t.Fatalf("expected a successful device flow authentication, got: %v", ctx.Error())
	}
	expected := "Please click the following link: https://example.com/device\n\nEnter the following code: ABCD-EFGH\n"
	if instruction != expected {
		t.Fatalf("unexpected device flow instruction: %q", instruction)
	}
	if ctx.Metadata().Username != "foo" {
		t.Fatalf("unexpected authenticated username: %q", ctx.Metadata().Username)
	}
}

func TestOAuth2KeyboardInteractiveDeviceFlowCustomPrompt(t *testing.T) {
	meta := testAuthPendingMetadata()
	flow := &fakeDeviceFlow{meta: meta}
	client := newTestOAuth2Client(
		t,
		&fakeOAuth2Provider{deviceFlow: flow},
		"Sign in at {{.AuthorizationURL}} with code {{.UserCode}}.",
		"",
	)

	var instruction string
	ctx := client.KeyboardInteractive(
		meta,
		func(i string, _ KeyboardInteractiveQuestions) (KeyboardInteractiveAnswers, error) {
			instruction = i
			return KeyboardInteractiveAnswers{}, nil
		},
	)
	if !ctx.Success() {
		t.Fatalf("expected a successful device flow authentication, got: %v", ctx.Error())
	}
	expected := "Sign in at https://example.com/device with code ABCD-EFGH."
	if instruction != expected {
		t.Fatalf("unexpected device flow instruction: %q", instruction)
	}
}

func TestOAuth2KeyboardInteractiveDeviceFlowPromptRenderFailure(t *testing.T) {
	meta := testAuthPendingMetadata()
	flow := &fakeDeviceFlow{meta: meta}
	// Parses fine, fails at render time on the real user code.
	client := newTestOAuth2Client(t, &fakeOAuth2Provider{deviceFlow: flow}, "{{index .UserCode 999}}", "")

	ctx := client.KeyboardInteractive(
		meta,
		func(_ string, _ KeyboardInteractiveQuestions) (KeyboardInteractiveAnswers, error) {
			t.Fatal("the challenge must not be sent if the prompt cannot be rendered")
			return KeyboardInteractiveAnswers{}, nil
		},
	)
	if ctx.Success() {
		t.Fatal("expected the authentication to fail when the prompt cannot be rendered")
	}
	if ctx.Error() == nil {
		t.Fatal("expected an error when the prompt cannot be rendered")
	}
}

func TestOAuth2KeyboardInteractiveDeviceFlowChallengeError(t *testing.T) {
	meta := testAuthPendingMetadata()
	flow := &fakeDeviceFlow{meta: meta}
	client := newTestOAuth2Client(t, &fakeOAuth2Provider{deviceFlow: flow}, "", "")

	ctx := client.KeyboardInteractive(
		meta,
		func(_ string, _ KeyboardInteractiveQuestions) (KeyboardInteractiveAnswers, error) {
			return KeyboardInteractiveAnswers{}, errors.New("client disconnected")
		},
	)
	if ctx.Success() {
		t.Fatal("expected the authentication to fail when the challenge errors")
	}
}

func TestOAuth2KeyboardInteractiveAuthorizationCodeFlow(t *testing.T) {
	meta := testAuthPendingMetadata()
	flow := &fakeAuthCodeFlow{meta: meta}
	client := newTestOAuth2Client(t, &fakeOAuth2Provider{authCodeFlow: flow}, "", "")

	var instruction string
	ctx := client.KeyboardInteractive(
		meta,
		func(i string, questions KeyboardInteractiveQuestions) (KeyboardInteractiveAnswers, error) {
			instruction = i
			if len(questions) != 1 {
				t.Fatalf("expected one question for the authorization code flow, got %d", len(questions))
			}
			return KeyboardInteractiveAnswers{
				Answers: map[string]string{"code": "the-state|the-code"},
			}, nil
		},
	)
	if !ctx.Success() {
		t.Fatalf("expected a successful authorization code flow authentication, got: %v", ctx.Error())
	}
	expected := "Please click the following link to log in: https://example.com/login\n\n"
	if instruction != expected {
		t.Fatalf("unexpected authorization code flow instruction: %q", instruction)
	}
	if flow.receivedState != "the-state" || flow.receivedCode != "the-code" {
		t.Fatalf("unexpected state/code: %q/%q", flow.receivedState, flow.receivedCode)
	}
}

func TestOAuth2KeyboardInteractiveAuthorizationCodeFlowInvalidCode(t *testing.T) {
	meta := testAuthPendingMetadata()
	flow := &fakeAuthCodeFlow{meta: meta}
	client := newTestOAuth2Client(t, &fakeOAuth2Provider{authCodeFlow: flow}, "", "")

	ctx := client.KeyboardInteractive(
		meta,
		func(_ string, _ KeyboardInteractiveQuestions) (KeyboardInteractiveAnswers, error) {
			return KeyboardInteractiveAnswers{
				Answers: map[string]string{"code": "missing-separator"},
			}, nil
		},
	)
	if ctx.Success() {
		t.Fatal("expected the authentication to fail for a code without state")
	}
}
