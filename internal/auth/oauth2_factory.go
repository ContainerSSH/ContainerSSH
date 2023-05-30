package auth

import (
	"fmt"
	"html/template"
	"io/ioutil"
	goHttp "net/http"

	"go.containerssh.io/libcontainerssh/config"
	"go.containerssh.io/libcontainerssh/http"
	"go.containerssh.io/libcontainerssh/internal/auth/oauth2"
	"go.containerssh.io/libcontainerssh/internal/metrics"
	"go.containerssh.io/libcontainerssh/log"
	"go.containerssh.io/libcontainerssh/message"
	"go.containerssh.io/libcontainerssh/service"
)

type oauthEndpointHandler struct {
	backend goHttp.FileSystem
	handler goHttp.Handler
}

type templatedata struct {
	Code string
}

func (o oauthEndpointHandler) ServeHTTP(writer goHttp.ResponseWriter, request *goHttp.Request) {
	if request.URL.Path != "/" {
		o.handler.ServeHTTP(writer, request)
		return
	}
	index, err := o.backend.Open("/index.html")
	if err != nil {
		writer.WriteHeader(500)
		panic(err)
	}
	defer func() {
		_ = index.Close()
	}()
	data, err := ioutil.ReadAll(index)
	if err != nil {
		writer.WriteHeader(500)
		panic(err)
	}
	tpl := template.New("index.html")
	tpl, err = tpl.Parse(string(data))
	if err != nil {
		panic(err)
	}
	query := request.URL.Query()
	state := query.Get("state")
	code := query.Get("code")
	writer.WriteHeader(200)
	writer.Header().Set("Content-Type", "text/html;charset=utf-8")
	if err := tpl.Execute(writer, templatedata{
		Code: state + "|" + code,
	}); err != nil {
		panic(err)
	}
}

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

	if _, err := fs.Open("index.html"); err != nil {
		return nil, nil, fmt.Errorf("oAuth redirect server does not contain an index.html file")
	}

	redirectServer, err := http.NewServer(
		"oAuth2 Redirect Server",
		cfg.Redirect.HTTPServerConfiguration,
		&oauthEndpointHandler{
			fs,
			goHttp.FileServer(fs),
		},
		logger,
		func(url string) {
			logger.Info(message.NewMessage(
				message.MAuthOAuth2Available,
				"oAuth2 redirect server is now available at %s",
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
	case config.AuthOAuth2OIDCProvider:
		provider, err = newOIDCProvider(cfg, logger)
		if err != nil {
			return nil, nil, err
		}
	case config.AuthOAuth2GenericProvider:
		provider, err = newGenericProvider(cfg, logger)
		if err != nil {
			return nil, nil, err
		}
	default:
		return nil, nil, message.NewMessage(
			message.EAuthConfigError,
			"Invalid oAuth2 provider: %s",
			cfg.Provider,
		)
	}

	return &oauth2Client{
		logger:   logger,
		provider: provider,
	}, redirectServer, nil
}
