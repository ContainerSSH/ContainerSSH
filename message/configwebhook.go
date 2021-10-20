package message

// ContainerSSH is sending a quest to the configuration server to obtain a per-user backend configuration.
const MConfigRequest = "CONFIG_REQUEST"

// ContainerSSH has received an invalid response from the configuration server or the network connection broke.
// ContainerSSH will retry fetching the user-specific configuration until the timeout. If this error persists check the
// connectivity to the configuration server, or the logs of the configuration server itself to find out of there may be
// a specific error.
const EConfigBackendError = "CONFIG_BACKEND_ERROR"

// ContainerSSH has received a per-user backend configuration from the configuration server.
const MConfigSuccess = "CONFIG_RESPONSE"

// ContainerSSH has received a non-200 response code when calling a per-user backend configuration from the
// configuration server.
const EConfigInvalidStatus = "CONFIG_INVALID_STATUS_CODE"

// The ContainerSSH configuration server is now available at the specified address.
const MConfigServerAvailable = "CONFIG_SERVER_AVAILABLE"

// The Listen option in the root configuration is deprecated since ContainerSSH 0.4. See https://containerssh.io/deprecations/listen for details.
const WConfigListenDeprecated = "CONFIG_LISTEN_DEPRECATED"
