package test

import (
    "context"
    "crypto/hmac"
    "crypto/sha256"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "html/template"
    "io"
    goHttp "net/http"
    "net/url"
    "strings"
    "testing"
    "time"

    "go.containerssh.io/libcontainerssh/config"
    "go.containerssh.io/libcontainerssh/http"
    "go.containerssh.io/libcontainerssh/log"
    "go.containerssh.io/libcontainerssh/service"
)

func OIDCServer(
    t *testing.T,
    redirectURI *url.URL,
) OIDCServerInstance {
    clientID := RandomString(6)
    clientSecret := RandomString(32)

    listen := fmt.Sprintf("127.0.0.1:%d", GetNextPort(t, "OIDC server"))
    handler := &oidcHandler{
        listen,
        clientID,
        clientSecret,
        map[string]string{},
        map[string]string{},
        map[string]string{},
        map[string]string{},
        redirectURI,
    }

    srv, err := http.NewServer(
        "OIDC server",
        config.HTTPServerConfiguration{
            Listen: listen,
        },
        handler,
        log.NewTestLogger(t),
        func(s string) {

        },
    )
    if err != nil {
        t.Fatal(err)
    }
    lifecycle := service.NewLifecycle(srv)
    t.Cleanup(func() {
        ctx, cancel := context.WithTimeout(context.Background(), time.Second)
        defer cancel()
        lifecycle.Stop(ctx)
        if err := lifecycle.Wait(); err != nil {
            t.Fatal(err)
        }
    })

    if err := lifecycle.Run(); err != nil {
        t.Fatal(err)
    }

    return &oidcServer{
        listen:    listen,
        lifecycle: lifecycle,
        handler:   handler,
    }
}

type OIDCServerInstance interface {
    BaseURL() string
    AuthorizationEndpoint() string
    TokenEndpoint() string
    UserinfoEndpoint() string
    ClientID() string
    ClientSecret() string
}

type oidcServer struct {
    lifecycle service.Lifecycle
    handler   *oidcHandler
    listen    string
}

func (o oidcServer) BaseURL() string {
    return o.listen
}

func (o oidcServer) AuthorizationEndpoint() string {
    return o.handler.AuthorizationEndpoint().String()
}

func (o oidcServer) DeviceEndpoint() string {
    return o.handler.DeviceEndpoint().String()
}

func (o oidcServer) TokenEndpoint() string {
    return o.handler.TokenEndpoint().String()
}

func (o oidcServer) UserinfoEndpoint() string {
    return o.handler.UserInfoEndpoint().String()
}

func (o oidcServer) ClientID() string {
    return o.handler.clientID
}

func (o oidcServer) ClientSecret() string {
    return o.handler.clientSecret
}

type oidcHandler struct {
    listen       string
    clientID     string
    clientSecret string
    users        map[string]string
    tokens       map[string]string
    authCodes    map[string]string
    deviceCodes  map[string]string
    redirectURI  *url.URL
}

func (o oidcHandler) ServeHTTP(writer goHttp.ResponseWriter, request *goHttp.Request) {
    switch strings.SplitN(request.RequestURI, "?", 2)[0] {
    case o.DiscoveryEndpoint().Path:
        o.configuration(writer, request)
    case o.DeviceEndpoint().Path:
        o.device(writer, request)
    case o.AuthorizationEndpoint().Path:
        o.authorization(writer, request)
    case o.TokenEndpoint().Path:
        o.token(writer, request)
    case o.UserInfoEndpoint().Path:
        o.userinfo(writer, request)
    }
}

func (o oidcHandler) DiscoveryEndpoint() *url.URL {
    result, err := url.Parse(o.listen + "/.well-known/openid-configuration")
    if err != nil {
        panic(err)
    }
    return result
}

func (o oidcHandler) AuthorizationEndpoint() *url.URL {
    result, err := url.Parse(o.listen + "/oauth/authorize")
    if err != nil {
        panic(err)
    }
    return result
}

func (o oidcHandler) TokenEndpoint() *url.URL {
    result, err := url.Parse(o.listen + "/oauth/token")
    if err != nil {
        panic(err)
    }
    return result
}

func (o oidcHandler) UserInfoEndpoint() *url.URL {
    result, err := url.Parse(o.listen + "/oauth/user")
    if err != nil {
        panic(err)
    }
    return result
}

func (o oidcHandler) RevocationEndpoint() *url.URL {
    result, err := url.Parse(o.listen + "/oauth/revoke")
    if err != nil {
        panic(err)
    }
    return result
}

func (o oidcHandler) JWKSEndpoint() *url.URL {
    result, err := url.Parse(o.listen + "/oauth/jwks")
    if err != nil {
        panic(err)
    }
    return result
}

func (o oidcHandler) DeviceEndpoint() *url.URL {
    result, err := url.Parse(o.listen + "/oauth/device")
    if err != nil {
        panic(err)
    }
    return result
}

var oidcFormTpl = template.Must(template.New("form").Parse(`
<!DOCTYPE html>
<html>
    <head><title>Login</title></head>
    <body>
        <form action="">
            <label for="username">Username:</label>
            <input type="text" id="username" name="username" />
            <label for="password">Username:</label>
            <input type="password" id="password" name="password" />
            <input type="submit" />
        </form>
    </body>
</html>`))

func (o *oidcHandler) authorization(writer goHttp.ResponseWriter, request *goHttp.Request) {
    u := request.URL
    queryValues := u.Query()
    responseType := queryValues.Get("response_type")
    switch responseType {
    case "code":
    default:
        o.sendError(writer, "unsupported_grant_type")
        return
    }
    if queryValues.Get("client_id") != o.clientID {
        o.sendError(writer, "unauthorized_client")
        return
    }
    redirectURI := o.redirectURI
    if queryValues.Get("redirect_uri") != "" {
        var err error
        redirectURI, err = url.Parse(queryValues.Get("redirect_uri"))
        if err != nil {
            o.sendError(writer, "invalid_request")
            return
        }
    }

    if request.Method == "POST" {
        body, err := io.ReadAll(request.Body)
        if err != nil {
            panic(err)
        }
        data, err := url.ParseQuery(string(body))
        if err != nil {
            writer.WriteHeader(400)
            return
        }

        username := data.Get("username")
        password := data.Get("password")

        if userPassword, ok := o.users[username]; ok {
            if userPassword == password {
                query := redirectURI.Query()
                query.Set("state", request.URL.Query().Get("state"))
                code := RandomString(16)
                o.authCodes[code] = username
                query.Set("code", code)
                redirectURI.RawQuery = query.Encode()
                writer.Header().Set("Location", redirectURI.String())
                return
            }
        }

        writer.Header().Set("Location", request.URL.String())
    }

    writer.Header().Set("Content-Type", "text/html")
    if err := oidcFormTpl.Execute(writer, nil); err != nil {
        panic(err)
    }
}

func (o *oidcHandler) sendError(writer goHttp.ResponseWriter, err string) {
    encoder := json.NewEncoder(writer)
    result := map[string]interface{}{
        "error": err,
    }
    writer.WriteHeader(400)
    writer.Header().Set("Content-Type", "application/json")
    if err := encoder.Encode(result); err != nil {
        panic(err)
    }
}

func (o *oidcHandler) token(writer goHttp.ResponseWriter, request *goHttp.Request) {
    body, err := io.ReadAll(request.Body)
    if err != nil {
        panic(err)
    }
    bodyData, err := url.ParseQuery(string(body))
    if err != nil {
        panic(err)
    }
    clientID, clientSecret := o.extractClientCredentials(bodyData, request)

    grantType := bodyData.Get("grant_type")

    var result map[string]interface{}
    switch grantType {
    case "authorization_code":
        if o.clientID != clientID || o.clientSecret != clientSecret {
            o.sendError(writer, "unauthorized_client")
            return
        }
        code := bodyData.Get("code")
        user, ok := o.authCodes[code]
        if !ok {
            o.sendError(writer, "invalid_grant")
            return
        }
        redirectURI := bodyData.Get("redirect_uri")
        if redirectURI == "" {
            o.sendError(writer, "invalid_request")
            return
        }
        delete(o.authCodes, code)
        result = o.issueToken(user, clientID)
    case "urn:ietf:params:oauth:grant-type:device_code":
        deviceCode := bodyData.Get("device_code")
        if o.clientID != clientID {
            o.sendError(writer, "unauthorized_client")
            return
        }
        user, ok := o.deviceCodes[deviceCode]
        if !ok {
            o.sendError(writer, "invalid_grant")
            return
        }
        delete(o.deviceCodes, deviceCode)
        result = o.issueToken(user, clientID)
    default:
        o.sendError(writer, "invalid_grant")
        return
    }
    writer.Header().Set("Content-Type", "application/json")
    encoder := json.NewEncoder(writer)
    if err := encoder.Encode(result); err != nil {
        panic(err)
    }
}

func (o *oidcHandler) extractClientCredentials(bodyData url.Values, request *goHttp.Request) (string, string) {
    clientID := bodyData.Get("client_id")
    clientSecret := bodyData.Get("client_secret")
    if clientID == "" && clientSecret == "" {
        authorization := strings.SplitN(request.Header.Get("Authorization"), " ", 2)
        if len(authorization) == 2 && strings.ToLower(authorization[0]) == "basic" {
            decodedAuthorization, err := base64.StdEncoding.DecodeString(authorization[1])
            if err == nil {
                authorizationParts := strings.SplitN(string(decodedAuthorization), ":", 2)
                if len(authorizationParts) == 2 {
                    clientID = authorizationParts[0]
                    clientSecret = authorizationParts[1]
                }
            }
        }
    }
    return clientID, clientSecret
}

func (o *oidcHandler) issueToken(user string, clientID string) map[string]interface{} {
    token := RandomString(16)
    o.tokens[token] = user
    idToken := map[string]interface{}{
        "iss": "https://example.com",
        "sub": user,
        "aud": clientID,
        "iat": time.Now().Unix(),
    }
    result := map[string]interface{}{}
    result["access_token"] = token
    result["token_type"] = "Bearer"
    jwt := createJWT(idToken, []byte(o.clientSecret))
    result["id_token"] = jwt
    return result
}

func createJWT(idToken map[string]interface{}, secret []byte) string {
    jwtHeader := map[string]interface{}{
        "typ": "JWT",
        "alg": "HS256",
    }
    headerData, err := json.Marshal(jwtHeader)
    if err != nil {
        panic(err)
    }
    encodedHeader := base64.RawURLEncoding.EncodeToString(headerData)
    payloadData, err := json.Marshal(idToken)
    if err != nil {
        panic(err)
    }
    encodedPayload := base64.RawURLEncoding.EncodeToString(payloadData)
    hm := hmac.New(sha256.New, secret)
    data := encodedHeader + "." + encodedPayload
    hm.Write([]byte(data))
    jwt := data + "." + base64.RawURLEncoding.EncodeToString(hm.Sum(nil))
    return strings.ReplaceAll(jwt, "=", "")
}

func (o *oidcHandler) userinfo(writer goHttp.ResponseWriter, request *goHttp.Request) {

}

func (o *oidcHandler) configuration(writer goHttp.ResponseWriter, request *goHttp.Request) {
    result := map[string]interface{}{
        "issuer":                 "https://example.com",
        "authorization_endpoint": o.AuthorizationEndpoint(),
        "token_endpoint":         o.TokenEndpoint(),
        "token_endpoint_auth_methods_supported": []string{
            "client_secret_basic",
            "client_secret_post",
        },
        "userinfo_endpoint":   o.UserInfoEndpoint(),
        "revocation_endpoint": o.RevocationEndpoint(),
        "revocation_endpoint_auth_methods_supported": []string{
            "client_secret_basic",
            "client_secret_post",
        },
        "device_authorization_endpoint": o.DeviceEndpoint(),
        "jwks_uri":                      o.JWKSEndpoint(),
        "response_types_supported":      []string{"code"},
        "grant_types_supported": []string{
            "authorization_code",
            "urn:ietf:params:oauth:grant-type:device_code",
        },
        "id_token_signing_alg_values_supported": []string{"HS256"},
    }
    o.writeJSONOutput(writer, result)
}

func (o *oidcHandler) writeJSONOutput(writer goHttp.ResponseWriter, result map[string]interface{}) {
    writer.Header().Set("Content-Type", "application/json")
    encoder := json.NewEncoder(writer)
    if err := encoder.Encode(result); err != nil {
        panic(err)
    }
}

func (o *oidcHandler) device(writer goHttp.ResponseWriter, request *goHttp.Request) {
    
}

func (o *oidcHandler) revocation(writer goHttp.ResponseWriter, request *goHttp.Request) {

}

func (o *oidcHandler) jwks(writer goHttp.ResponseWriter, request *goHttp.Request) {
    clientID, clientSecret := o.extractClientCredentials(url.Values{}, request)
    if o.clientID != clientID || o.clientSecret != clientSecret {
        o.sendError(writer, "unauthorized_client")
        return
    }
    result := map[string]interface{}{
        "keys": []map[string]interface{}{
            {
                "alg": "HS256",
                "kty": "oct",
                "use": "sig",
                "k":   o.clientSecret,
            },
        },
    }
    o.writeJSONOutput(writer, result)
}
