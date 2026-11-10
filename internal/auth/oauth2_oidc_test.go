package auth_test

import (
	"encoding/json"
	"errors"
	goHttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.containerssh.io/containerssh/config"
	"go.containerssh.io/containerssh/internal/auth"
	"go.containerssh.io/containerssh/internal/geoip/dummy"
	"go.containerssh.io/containerssh/internal/metrics"
	"go.containerssh.io/containerssh/internal/structutils"
	"go.containerssh.io/containerssh/log"
	"go.containerssh.io/containerssh/message"
	"go.containerssh.io/containerssh/metadata"
)

func TestOIDCEnforceUsername(t *testing.T) {
	for _, tc := range []struct {
		name            string
		sshUsername     string
		enforceUsername bool
		success         bool
	}{
		{"enforced, matching username", "foo", true, true},
		{"enforced, different username", "bar", true, false},
		{"not enforced, different username", "bar", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			logger := log.NewTestLogger(t)
			oidcServer := newFakeOIDCServer(t, map[string]interface{}{
				"sub":                "0123456789",
				"preferred_username": "foo",
			})

			cfg := config.AuthOAuth2ClientConfig{}
			structutils.Defaults(&cfg)
			cfg.ClientID = "containerssh"
			cfg.ClientSecret = "secret"
			cfg.Provider = config.AuthOAuth2OIDCProvider
			cfg.OIDC.URL = oidcServer.URL + "/"
			cfg.OIDC.DeviceFlow = true
			cfg.OIDC.UsernameField = "preferred_username"
			cfg.OIDC.EnforceUsername = tc.enforceUsername

			client, _, err := auth.NewOAuth2Client(cfg, logger, metrics.New(dummy.New()))
			if err != nil {
				t.Fatalf("failed to create OAuth2 client (%v)", err)
			}

			authContext := client.KeyboardInteractive(
				metadata.NewTestAuthenticatingMetadata(tc.sshUsername),
				func(
					_ string,
					_ auth.KeyboardInteractiveQuestions,
				) (auth.KeyboardInteractiveAnswers, error) {
					return auth.KeyboardInteractiveAnswers{}, nil
				},
			)
			defer authContext.OnDisconnect()

			assert.Equal(t, tc.success, authContext.Success())
			if tc.success {
				assert.NoError(t, authContext.Error())
				assert.Equal(t, "foo", authContext.Metadata().AuthenticatedUsername)
				return
			}
			var msg message.Message
			if !errors.As(authContext.Error(), &msg) {
				t.Fatalf("unexpected error type %T (%v)", authContext.Error(), authContext.Error())
			}
			assert.Equal(t, message.EAuthOIDCUsernameDoesNotMatch, msg.Code())
		})
	}
}

// newFakeOIDCServer starts an OIDC server that immediately approves device flow logins and returns the passed user
// info.
func newFakeOIDCServer(t *testing.T, userInfo map[string]interface{}) *httptest.Server {
	mux := goHttp.NewServeMux()
	var srv *httptest.Server
	writeJSON := func(w goHttp.ResponseWriter, data interface{}) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(data); err != nil {
			t.Errorf("failed to encode response (%v)", err)
		}
	}
	mux.HandleFunc("/.well-known/openid-configuration", func(w goHttp.ResponseWriter, _ *goHttp.Request) {
		writeJSON(w, map[string]string{
			"token_endpoint":                srv.URL + "/token",
			"userinfo_endpoint":             srv.URL + "/userinfo",
			"device_authorization_endpoint": srv.URL + "/device",
			"revocation_endpoint":           srv.URL + "/revoke",
		})
	})
	mux.HandleFunc("/device", func(w goHttp.ResponseWriter, _ *goHttp.Request) {
		writeJSON(w, map[string]interface{}{
			"device_code":      "device-code",
			"user_code":        "user-code",
			"verification_uri": srv.URL + "/verify",
			"expires_in":       60,
			"interval":         1,
		})
	})
	mux.HandleFunc("/token", func(w goHttp.ResponseWriter, _ *goHttp.Request) {
		writeJSON(w, map[string]string{
			"access_token": "access-token",
			"token_type":   "Bearer",
			"scope":        "openid",
		})
	})
	mux.HandleFunc("/userinfo", func(w goHttp.ResponseWriter, r *goHttp.Request) {
		if r.Header.Get("Authorization") != "Bearer access-token" {
			w.WriteHeader(goHttp.StatusUnauthorized)
			return
		}
		writeJSON(w, userInfo)
	})
	mux.HandleFunc("/revoke", func(w goHttp.ResponseWriter, _ *goHttp.Request) {
		w.WriteHeader(goHttp.StatusOK)
	})
	srv = httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}
