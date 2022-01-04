package auth

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/containerssh/libcontainerssh/auth"
	"github.com/containerssh/libcontainerssh/log"
	"github.com/containerssh/libcontainerssh/message"
)

type oauth2Client struct {
	provider OAuth2Provider
	logger   log.Logger
}

type oauth2Context struct {
	success  bool
	metadata *auth.ConnectionMetadata
	err      error
	flow     OAuth2Flow
}

func (o *oauth2Context) Success() bool {
	return o.success
}

func (o *oauth2Context) Error() error {
	return o.err
}

func (o *oauth2Context) Metadata() *auth.ConnectionMetadata {
	return o.metadata
}

func (o *oauth2Context) OnDisconnect() {
	if o.flow != nil {
		o.flow.Deauthorize(context.TODO())
	}
}

func (o *oauth2Client) Password(_ string, _ []byte, _ string, _ net.IP) AuthenticationContext {
	return &oauth2Context{false, nil, message.UserMessage(
		message.EAuthUnsupported,
		"Password authentication is not available.",
		"OAuth2 doesn't support password authentication.",
	), nil}
}

func (o *oauth2Client) PubKey(_ string, _ string, _ string, _ net.IP) AuthenticationContext {
	return &oauth2Context{false, nil, message.UserMessage(
		message.EAuthUnsupported,
		"Public key authentication is not available.",
		"OAuth2 doesn't support public key authentication.",
	), nil}
}

func (client *oauth2Client) GSSAPIConfig(connectionId string, addr net.IP) GSSAPIServer {
	return nil
}

func (o *oauth2Client) KeyboardInteractive(
	username string,
	challenge func(
		instruction string,
		questions KeyboardInteractiveQuestions,
	) (
		answers KeyboardInteractiveAnswers,
		err error,
	),
	connectionID string,
	_ net.IP,
) AuthenticationContext {
	ctx := context.TODO()
	metadata := &auth.ConnectionMetadata{}
	var err error
	if o.provider.SupportsDeviceFlow() {
		deviceFlow, err := o.provider.GetDeviceFlow(connectionID, username)
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
					return &oauth2Context{false, metadata, err, deviceFlow}
				}
				verifyContext, cancelFunc := context.WithTimeout(ctx, expiration)
				defer cancelFunc()
				metadata.Metadata, err = deviceFlow.Verify(verifyContext)
				// TODO fallback to authorization code flow if the device flow rate limit is exceeded.
				if err != nil {
					deviceFlow.Deauthorize(ctx)
					return &oauth2Context{false, metadata, err, deviceFlow}
				} else {
					return &oauth2Context{true, metadata, nil, deviceFlow}
				}
			}
		}
	}
	if o.provider.SupportsAuthorizationCodeFlow() {
		authCodeFlow, err := o.provider.GetAuthorizationCodeFlow(connectionID, username)
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
					return &oauth2Context{false, metadata, err, authCodeFlow}
				} else {
					if code, ok := answers.Answers["code"]; ok {
						parts := strings.SplitN(code, "|", 2)
						if len(parts) != 2 {
							return &oauth2Context{false, metadata, message.UserMessage(
								message.EAuthFailed,
								"Authentication failed.",
								"Authentication failed because the return code did not contain the requisite state and code.",
							), authCodeFlow}
						}
						metadata.Metadata, err = authCodeFlow.Verify(ctx, parts[0], parts[1])
						if err != nil {
							return &oauth2Context{false, metadata, err, authCodeFlow}
						} else {
							return &oauth2Context{true, metadata, nil, authCodeFlow}
						}
					} else {
						return &oauth2Context{false, metadata, err, authCodeFlow}
					}
				}
			}
		}
	}
	return &oauth2Context{false, metadata, message.WrapUser(err,
		message.EAuthFailed, "Authentication failed.", "Authentication failed."), nil}
}
