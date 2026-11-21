package message

// ECoreConfig indicates that ContainerSSH encountered an error in the configuration.
const ECoreConfig = "CORE_CONFIG_ERROR"

// ECoreNoHostKeys indicates that the configuration does not contain host keys. ContainerSSH will attempt to generate host keys and update the configuration file.
const ECoreNoHostKeys = "CORE_NO_HOST_KEYS"

// ECoreHostKeyGenerationFailed indicates that ContainerSSH could not generate host keys and is aborting the run.
const ECoreHostKeyGenerationFailed = "CORE_HOST_KEY_GENERATION_FAILED"

// ECannotWriteConfigFile indicates that ContainerSSH cannot update the configuration file with the new host keys and will only use the host key for the current run.
const ECannotWriteConfigFile = "CORE_CONFIG_CANNOT_WRITE_FILE"

// MCoreConfigFile indicates that ContainerSSH is reading the configuration file.
const MCoreConfigFile = "CORE_CONFIG_FILE"

// ECoreHealthCheckFailed indicates that a ContainerSSH health check failed.
const ECoreHealthCheckFailed = "CORE_HEALTH_CHECK_FAILED"

// MCoreHealthCheckSuccessful indicates that The health check was successful.
const MCoreHealthCheckSuccessful = "CORE_HEALTH_CHECK_SUCCESSFUL"
