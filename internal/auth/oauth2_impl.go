package auth

import (
	"context"
	"strings"
	"text/template"
	"time"

	"go.containerssh.io/containerssh/config"
	"go.containerssh.io/containerssh/log"
	"go.containerssh.io/containerssh/message"
	"go.containerssh.io/containerssh/metadata"
)

type oauth2Client struct {
	provider                    OAuth2Provider
	logger                      log.Logger
	deviceFlowPrompt            *template.Template
	authorizationCodeFlowPrompt *template.Template
}

func renderOAuth2Prompt(tpl *template.Template, data interface{}) (string, error) {
	prompt := &strings.Builder{}
	if err := tpl.Execute(prompt, data); err != nil {
		return "", message.WrapUser(
			err,
			message.EAuthConfigError,
			"Authentication failed.",
			"Failed to render the oAuth2 %s template.",
			tpl.Name(),
		)
	}
	return prompt.String(), nil
}

type oauth2Context struct {
	success  bool
	metadata metadata.ConnectionAuthenticatedMetadata
	err      error
	flow     OAuth2Flow
}

func (o *oauth2Context) Success() bool {
	return o.success
}

func (o *oauth2Context) Error() error {
	return o.err
}

func (o *oauth2Context) Metadata() metadata.ConnectionAuthenticatedMetadata {
	return o.metadata
}

func (o *oauth2Context) OnDisconnect() {
	if o.flow != nil {
		// TODO proper timeout handling
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		o.flow.Deauthorize(ctx)
	}
}

func (o *oauth2Client) KeyboardInteractive(
	meta metadata.ConnectionAuthPendingMetadata,
	challenge func(
		instruction string,
		questions KeyboardInteractiveQuestions,
	) (
		answers KeyboardInteractiveAnswers,
		err error,
	),
) AuthenticationContext {
	ctx := context.TODO()
	var err error
	if o.provider.SupportsDeviceFlow() {
		deviceFlow, err := o.provider.GetDeviceFlow(ctx, meta)
		if err == nil {
			authorizationURL, userCode, expiration, err := deviceFlow.GetAuthorizationURL(ctx)
			if err == nil {
				prompt, promptErr := renderOAuth2Prompt(
					o.deviceFlowPrompt,
					config.OAuth2DeviceFlowPromptData{
						AuthorizationURL: authorizationURL,
						UserCode:         userCode,
					},
				)
				if promptErr != nil {
					return &oauth2Context{false, meta.AuthFailed(), promptErr, deviceFlow}
				}
				_, err = challenge(
					prompt,
					KeyboardInteractiveQuestions{},
				)
				if err != nil {
					return &oauth2Context{false, meta.AuthFailed(), err, deviceFlow}
				}
				verifyContext, cancelFunc := context.WithTimeout(ctx, expiration)
				defer cancelFunc()
				_, authenticatedMeta, err := deviceFlow.Verify(verifyContext)
				// TODO fallback to authorization code flow if the device flow rate limit is exceeded.
				if err != nil {
					deviceFlow.Deauthorize(ctx)
					return &oauth2Context{false, authenticatedMeta, err, deviceFlow}
				} else {
					return &oauth2Context{true, authenticatedMeta, nil, deviceFlow}
				}
			}
		}
	}
	if o.provider.SupportsAuthorizationCodeFlow() {
		authCodeFlow, err := o.provider.GetAuthorizationCodeFlow(ctx, meta)
		if err == nil {
			link, err := authCodeFlow.GetAuthorizationURL(ctx)
			if err == nil {
				prompt, promptErr := renderOAuth2Prompt(
					o.authorizationCodeFlowPrompt,
					config.OAuth2AuthorizationCodeFlowPromptData{
						AuthorizationURL: link,
					},
				)
				if promptErr != nil {
					return &oauth2Context{false, meta.AuthFailed(), promptErr, authCodeFlow}
				}
				answers, err := challenge(
					prompt,
					KeyboardInteractiveQuestions{
						KeyboardInteractiveQuestion{
							ID:           "code",
							Question:     "Please paste the received code: ",
							EchoResponse: false,
						},
					},
				)
				if err != nil {
					return &oauth2Context{false, meta.AuthFailed(), err, authCodeFlow}
				} else {
					if code, ok := answers.Answers["code"]; ok {
						parts := strings.SplitN(code, "|", 2)
						if len(parts) != 2 {
							return &oauth2Context{
								false, meta.AuthFailed(), message.UserMessage(
									message.EAuthFailed,
									"Authentication failed.",
									"Authentication failed because the return code did not contain the requisite state and code.",
								), authCodeFlow,
							}
						}
						_, authenticatedMeta, err := authCodeFlow.Verify(ctx, parts[0], parts[1])
						if err != nil {
							return &oauth2Context{false, authenticatedMeta, err, authCodeFlow}
						} else {
							return &oauth2Context{true, authenticatedMeta, nil, authCodeFlow}
						}
					} else {
						return &oauth2Context{false, meta.AuthFailed(), err, authCodeFlow}
					}
				}
			}
		}
	}
	return &oauth2Context{
		false, meta.AuthFailed(), message.WrapUser(
			err,
			message.EAuthFailed, "Authentication failed.", "Authentication failed.",
		), nil,
	}
}
