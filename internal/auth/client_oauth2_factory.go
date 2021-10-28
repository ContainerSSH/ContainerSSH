package auth

import (
	"fmt"
	goHttp "net/http"

	"github.com/containerssh/libcontainerssh/config"
	"github.com/containerssh/libcontainerssh/http"
	"github.com/containerssh/libcontainerssh/internal/auth/oauth2"
	"github.com/containerssh/libcontainerssh/internal/metrics"
	"github.com/containerssh/libcontainerssh/log"
	"github.com/containerssh/libcontainerssh/message"
	"github.com/containerssh/libcontainerssh/service"
)

func NewOAuth2Client(cfg config.AuthConfig, logger log.Logger, collector metrics.Collector) (
	Client,
	service.Service,
	error,
) {
	var err error
	if cfg.Method != config.AuthMethodOAuth2 {
		return nil, nil, fmt.Errorf("authentication is not set to oauth2")
	}
	if err := cfg.Validate(); err != nil {
		return nil, nil, err
	}

	var fs goHttp.FileSystem
	if cfg.OAuth2.Redirect.Webroot != "" {
		fs = goHttp.Dir(cfg.OAuth2.Redirect.Webroot)
	} else {
		fs = oauth2.GetFilesystem()
	}

	redirectServer, err := http.NewServer(
		"OAuth2 Redirect Server",
		cfg.OAuth2.Redirect.HTTPServerConfiguration,
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
	switch cfg.OAuth2.Provider {
	case config.AuthOAuth2GitHubProvider:
		provider, err = newGitHubProvider(cfg, logger)
	}

	return &oauth2Client{
		logger:   logger,
		provider: provider,
	}, redirectServer, nil
}
