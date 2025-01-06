package message

// MServiceStarting indicates that ContainerSSH is starting a component service.
const MServiceStarting = "SERVICE_STARTING"

// MServiceRunning indicates that a ContainerSSH service is now running.
const MServiceRunning = "SERVICE_RUNNING"

// MServiceStopping indicates that a ContainerSSH service is now stopping.
const MServiceStopping = "SERVICE_STOPPING"

// MServiceStopped indicates that a ContainerSSH service has stopped.
const MServiceStopped = "SERVICE_STOPPED"

// EServiceCrashed indicates that a ContainerSSH has stopped improperly.
const EServiceCrashed = "SERVICE_CRASHED"

// MServicesStarting indicates that all ContainerSSH services are starting.
const MServicesStarting = "SERVICE_POOL_STARTING"

// MServicesRunning indicates that all ContainerSSH services are now running.
const MServicesRunning = "SERVICE_POOL_RUNNING"

// MServicesStopping indicates that ContainerSSH is stopping all services.
const MServicesStopping = "SERVICE_POOL_STOPPING"

// MServicesStopped indicates that ContainerSSH has stopped all services.
const MServicesStopped = "SERVICE_POOL_STOPPED"
