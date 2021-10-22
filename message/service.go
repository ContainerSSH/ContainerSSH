package message

// ContainerSSH is starting a component service
const MServiceStarting = "SERVICE_STARTING"

// A ContainerSSH service is now running
const MServiceRunning = "SERVICE_RUNNING"

// A ContainerSSH service is now stopping.
const MServiceStopping = "SERVICE_STOPPING"

// A ContainerSSH service has stopped.
const MServiceStopped = "SERVICE_STOPPED"

// A ContainerSSH has stopped improperly.
const EServiceCrashed = "SERVICE_CRASHED"

// All ContainerSSH services are starting.
const MServicesStarting = "SERVICE_POOL_STARTING"

// All ContainerSSH services are now running.
const MServicesRunning = "SERVICE_POOL_RUNNING"

// ContainerSSH is stopping all services.
const MServicesStopping = "SERVICE_POOL_STOPPING"

// ContainerSSH has stopped all services.
const MServicesStopped = "SERVICE_POOL_STOPPED"
