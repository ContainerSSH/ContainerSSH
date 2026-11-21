# Message / error codes

| Code | Explanation |
|------|-------------|
| `CORE_CONFIG_CANNOT_WRITE_FILE` | ContainerSSH cannot update the configuration file with the new host keys and will only use the host key for the current run. |
| `CORE_CONFIG_ERROR` | ContainerSSH encountered an error in the configuration. |
| `CORE_CONFIG_FILE` | ContainerSSH is reading the configuration file. |
| `CORE_HEALTH_CHECK_FAILED` | A ContainerSSH health check failed. |
| `CORE_HEALTH_CHECK_SUCCESSFUL` | The health check was successful. |
| `CORE_HOST_KEY_GENERATION_FAILED` | ContainerSSH could not generate host keys and is aborting the run. |
| `CORE_NO_HOST_KEYS` | The configuration does not contain host keys. ContainerSSH will attempt to generate host keys and update the configuration file. |

