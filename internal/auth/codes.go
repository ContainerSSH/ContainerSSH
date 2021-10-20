package auth

// This message indicates that the authentication server returned an invalid HTTP status code.
const EInvalidStatus = "AUTH_INVALID_STATUS"

// ContainerSSH is trying to contact the authentication backend to verify the user credentials.
const MAuth = "AUTH"

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
const ERequestDecodeFailed = "AUTH_SERVER_DECODE_FAILED"

// The authentication method the client tried is disabled.
const EDisabled = "AUTH_DISABLED"

// The authentication the client requested is not available for the current method.
const EUnsupported = "AUTH_UNSUPPORTED"

// You are using a deprecated feature and should consider switching.
const EDeprecated = "AUTH_DEPRECATED"

// This message indicates that the OAuth2 redirect server is now running at the specified address.
const EOAuth2Available = "AUTH_OAUTH2_AVAILABLE"

// ContainerSSH failed to fetch the access token after the authentication process finishes.
const EGitHubAccessTokenFetchFailed = "GITHUB_ACCESS_TOKEN_FETCH_FAILED"

// ContainerSSH is still waiting for the user to enter the device code and authorize ContainerSSH.
const EGitHubAuthorizationPending = "GITHUB_AUTHORIZATION_PENDING"

// Authentication via GitHub timed out.
const EGitHubTimeout = "GITHUB_TIMEOUT"

// The user did not grand a required scope permission on GitHub.
const EGitHubRequiredScopeNotGranted = "GITHUB_REQUIRED_SCOPE_NOT_GRANTED"

// The user does not have two factor authentication enabled on GitHub, but it is required in the configuration.
const EGitHubNo2FA = "GITHUB_2FA_NOT_ENABLED"

// Fetching the user information from GitHub failed.
const EGitHubUserRequestFailed = "GITHUB_USER_REQUEST_FAILED"

// The username entered in the SSH login and the GitHub login name do not match, but enforceUsername is enabled. This is
// done as a safety mechanism, otherwise all other components would assume the SSH login is the correct one. It can
// be disabled and let the configuration server send a custom configuration based on the GITHUB_LOGIN metadata entry.
const EUsernameDoesNotMatch = "GITHUB_USER_DOES_NOT_MATCH"

// The user provided a return code that contained an invalid state component. This usually means that the user did not
// copy the code correctly, but may also mean that the code was manipulated.
const EGitHubStateDoesNotMatch = "GITHUB_STATE_DOES_NOT_MATCH"

// The GitHub device authorization count per hour (currently: 50) has been reached. ContainerSSH is falling back to the
// authorization code flow for this authentication.
const EGitHubDeviceAuthorizationLimit = "GITHUB_DEVICE_AUTHORIZATION_LIMIT"

// ContainerSSH failed to delete the GitHub access token when the user logged out.
const EGitHubDeleteAccessTokenFailed = "GITHUB_DELETE_ACCESS_TOKEN_FAILED"

// ContainerSSH failed to fetch a device code for the device authentication flow.
const EGitHubDeviceCodeRequestFailed = "GITHUB_DEVICE_CODE_REQUEST_FAILED"

// ContainerSSH failed to create a HTTP client for communicating with GitHub. This is likely a bug, please report it.
const EGitHubHTTPClientCreateFailed = "GITHUB_HTTP_CLIENT_CREATE_FAILED"

// ContainerSSH has detected that the OAuth client ID or secret are invalid.
const EIncorrectClientCredentials = "OAUTH_INCORRECT_CLIENT_CREDENTIALS"

// The device flow rate limit on the OAuth2 server has been exceeded, falling back to authorization code flow.
const EDeviceFlowRateLimitExceeded = "OAUTH_DEVICE_FLOW_RATE_LIMIT_EXCEEDED"

// This error means that the user specified a username other than their GitHub login and enforceUsername was set to on.
const EGitHubUsernameDoesNotMatch = "GITHUB_USERNAME_DOES_NOT_MATCH"