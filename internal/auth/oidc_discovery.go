package auth

import (
	"context"
	"time"

	"go.containerssh.io/libcontainerssh/http"
	"go.containerssh.io/libcontainerssh/log"
	"go.containerssh.io/libcontainerssh/message"
)

func newOIDCDiscovery(logger log.Logger) oidcDiscovery {
	return &oidcDiscoverImpl{
		logger: logger,
	}
}

type oidcDiscovery interface {
	Discover(ctx context.Context, httpClient http.Client) (oidcDiscoveryResponse, error)
}

type oidcDiscoveryResponse struct {
	AuthorizationEndpoint       string `json:"authorization_endpoint"`
	TokenEndpoint               string `json:"token_endpoint"`
	UserInfoEndpoint            string `json:"userinfo_endpoint"`
	DeviceAuthorizationEndpoint string `json:"device_authorization_endpoint"`
	RevocationEndpoint          string `json:"revocation_endpoint,omitempty"`
}

type oidcDiscoverImpl struct {
	//httpClient http.Client
	logger log.Logger
}

func (o oidcDiscoverImpl) Discover(ctx context.Context, httpClient http.Client) (oidcDiscoveryResponse, error) {
	var statusCode int
	var err error
	for {
		response := oidcDiscoveryResponse{}
		statusCode, err = httpClient.Get(".well-known/openid-configuration", &response)
		if err == nil && statusCode == 200 {
			return response, nil
		}
		err = message.WrapUser(
			err,
			message.EAuthOIDCDiscoveryFailed,
			"Authentication currently unavailable",
			"OIDC endpoint configuration request failed.",
		)
		o.logger.Debug(err)
		if statusCode > 399 && statusCode < 500 {
			return oidcDiscoveryResponse{}, err
		}

		select {
		case <-time.After(5 * time.Second):
		case <-ctx.Done():
			err = message.WrapUser(
				err,
				message.EAuthOIDCDiscoveryTimeout,
				"Authentication currently unavailable",
				"Timeout while performing OIDC discovery.",
			)
			o.logger.Warning(err)
			return oidcDiscoveryResponse{}, err
		}
	}
}
