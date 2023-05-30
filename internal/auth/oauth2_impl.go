package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.containerssh.io/libcontainerssh/log"
	"go.containerssh.io/libcontainerssh/message"
	"go.containerssh.io/libcontainerssh/metadata"
)

type oauth2Client struct {
	provider OAuth2Provider
	logger   log.Logger
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
				_, err = challenge(
					fmt.Sprintf(
						"Please click the following link: %s\n\nEnter the following code: %s\n",
						authorizationURL,
						userCode,
					),
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
				answers, err := challenge(
					fmt.Sprintf(
						"Please click the following link to log in: %s\n\n",
						link,
					),
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
