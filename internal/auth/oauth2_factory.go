package auth

import (
	"fmt"
	goHttp "net/http"

    "go.containerssh.io/libcontainerssh/config"
    "go.containerssh.io/libcontainerssh/http"
    "go.containerssh.io/libcontainerssh/internal/auth/oauth2"
    "go.containerssh.io/libcontainerssh/internal/metrics"
    "go.containerssh.io/libcontainerssh/log"
    "go.containerssh.io/libcontainerssh/message"
    "go.containerssh.io/libcontainerssh/service"
)

func NewOAuth2Client(cfg config.AuthOAuth2ClientConfig, logger log.Logger, collector metrics.Collector) (
	KeyboardInteractiveAuthenticator,
	service.Service,
	error,
) {
	var err error
	if err := cfg.Validate(); err != nil {
		return nil, nil, err
	}

	var fs goHttp.FileSystem
	if cfg.Redirect.Webroot != "" {
		fs = goHttp.Dir(cfg.Redirect.Webroot)
	} else {
		fs = oauth2.GetFilesystem()
	}

	redirectServer, err := http.NewServer(
		"OAuth2 Redirect Server",
		cfg.Redirect.HTTPServerConfiguration,
		goHttp.FileServer(fs),
		logger,
		func(url string) {
			logger.Info(message.NewMessage(
				message.EAuthOAuth2Available,
				"OAuth2 redirect server is now available at %s",
				url,
			))
		},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create redirect page server (%w)", err)
	}

	var provider OAuth2Provider
	switch cfg.Provider {
	case config.AuthOAuth2GitHubProvider:
		provider, err = newGitHubProvider(cfg, logger)
		if err != nil {
			return nil, nil, err
		}
	}

	return &oauth2Client{
		logger:   logger,
		provider: provider,
	}, redirectServer, nil
}
