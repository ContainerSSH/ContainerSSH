package message

// This message indicates that the authentication server returned an invalid HTTP status code.
const EAuthInvalidStatus = "AUTH_INVALID_STATUS"

// ContainerSSH is trying to contact the authentication backend to verify the user credentials.
const MAuth = "AUTH"

// The authentication configuration is invalid.
const EAuthConfigError = "AUTH_CONFIG_ERROR"

// The ContainerSSH authentication server responded with a non-200 status code. ContainerSSH will retry the
// authentication for a few times before giving up. This is most likely a bug in your authentication server, please
// check your logs.
const EAuthBackendError = "AUTH_BACKEND_ERROR"

// The user has provided invalid credentials and the authentication is rejected.
const EAuthFailed = "AUTH_FAILED"

// The user has provided the correct credentials and the authentication is accepted.
const MAuthSuccessful = "AUTH_SUCCESSFUL"

// The ContainerSSH authentication server is now available.
const MAuthServerAvailable = "AUTH_AVAILABLE"

// The ContainerSSH Auth library failed to decode a request from ContainerSSH.
const EAuthRequestDecodeFailed = "AUTH_SERVER_DECODE_FAILED"

// The authentication method the client tried is disabled.
const EAuthDisabled = "AUTH_DISABLED"

// The authentication the client requested is not available for the current method.
const EAuthUnsupported = "AUTH_UNSUPPORTED"

// You are using a deprecated feature and should consider switching.
const EAuthDeprecated = "AUTH_DEPRECATED"

// This message indicates that the OAuth2 redirect server is now running at the specified address.
const EAuthOAuth2Available = "AUTH_OAUTH2_AVAILABLE"

// ContainerSSH failed to fetch the access token after the authentication process finishes.
const EAuthGitHubAccessTokenFetchFailed = "GITHUB_ACCESS_TOKEN_FETCH_FAILED"

// ContainerSSH is still waiting for the user to enter the device code and authorize ContainerSSH.
const EAuthGitHubAuthorizationPending = "GITHUB_AUTHORIZATION_PENDING"

// Authentication via GitHub timed out.
const EAuthGitHubTimeout = "GITHUB_TIMEOUT"

// The user did not grand a required scope permission on GitHub.
const EAuthGitHubRequiredScopeNotGranted = "GITHUB_REQUIRED_SCOPE_NOT_GRANTED"

// The user does not have two factor authentication enabled on GitHub, but it is required in the configuration.
const EAuthGitHubNo2FA = "GITHUB_2FA_NOT_ENABLED"

// Fetching the user information from GitHub failed.
const EAuthGitHubUserRequestFailed = "GITHUB_USER_REQUEST_FAILED"

// The username entered in the SSH login and the GitHub login name do not match, but enforceUsername is enabled. This is
// done as a safety mechanism, otherwise all other components would assume the SSH login is the correct one. It can
// be disabled and let the configuration server send a custom configuration based on the GITHUB_LOGIN metadata entry.
const EAuthUsernameDoesNotMatch = "GITHUB_USER_DOES_NOT_MATCH"

// The user provided a return code that contained an invalid state component. This usually means that the user did not
// copy the code correctly, but may also mean that the code was manipulated.
const EAuthGitHubStateDoesNotMatch = "GITHUB_STATE_DOES_NOT_MATCH"

// The GitHub device authorization count per hour (currently: 50) has been reached. ContainerSSH is falling back to the
// authorization code flow for this authentication.
const EAuthGitHubDeviceAuthorizationLimit = "GITHUB_DEVICE_AUTHORIZATION_LIMIT"

// ContainerSSH failed to delete the GitHub access token when the user logged out.
const EAuthGitHubDeleteAccessTokenFailed = "GITHUB_DELETE_ACCESS_TOKEN_FAILED"

// ContainerSSH failed to fetch a device code for the device authentication flow.
const EAuthGitHubDeviceCodeRequestFailed = "GITHUB_DEVICE_CODE_REQUEST_FAILED"

// ContainerSSH failed to create a HTTP client for communicating with GitHub. This is likely a bug, please report it.
const EAuthGitHubHTTPClientCreateFailed = "GITHUB_HTTP_CLIENT_CREATE_FAILED"

// ContainerSSH has detected that the OAuth client ID or secret are invalid.
const EAuthIncorrectClientCredentials = "OAUTH_INCORRECT_CLIENT_CREDENTIALS"

// The device flow rate limit on the OAuth2 server has been exceeded, falling back to authorization code flow.
const EAuthDeviceFlowRateLimitExceeded = "OAUTH_DEVICE_FLOW_RATE_LIMIT_EXCEEDED"

// This error means that the user specified a username other than their GitHub login and enforceUsername was set to on.
const EAuthGitHubUsernameDoesNotMatch = "GITHUB_USERNAME_DOES_NOT_MATCH"
