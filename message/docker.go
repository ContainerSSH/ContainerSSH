package message

// EDockerFailedAgentRead indicates that the ContainerSSH Docker module failed to read from the ContainerSSH agent. This
// is most likely because the ContainerSSH guest agent is not present in the guest image, but agent support is
// enabled.
const EDockerFailedAgentRead = "DOCKER_AGENT_READ_FAILED"

// EDockerFailedOutputCloseWriting indicates that the ContainerSSH Docker module attempted to close the output (stdout and
// stderr) for writing but failed to do so.
const EDockerFailedOutputCloseWriting = "DOCKER_CLOSE_OUTPUT_FAILED"

// EDockerFailedInputCloseWriting indicates that the ContainerSSH Docker module attempted to close the input (stdin) for
// reading but failed to do so.
const EDockerFailedInputCloseWriting = "DOCKER_CLOSE_INPUT_FAILED"

// EDockerConfigError indicates that the ContainerSSH Docker module detected a configuration error. Please check your
// configuration.
const EDockerConfigError = "DOCKER_CONFIG_ERROR"

// MDockerContainerAttach indicates that the ContainerSSH Docker module is attaching to a container in session mode.
const MDockerContainerAttach = "DOCKER_CONTAINER_ATTACH"

// EDockerFailedContainerAttach indicates that the ContainerSSH Docker module has failed to attach to a container in
// session mode.
const EDockerFailedContainerAttach = "DOCKER_CONTAINER_ATTACH_FAILED"

// MDockerContainerCreate indicates that the ContainerSSH Docker module is creating a container.
const MDockerContainerCreate = "DOCKER_CONTAINER_CREATE"

// EDockerFailedContainerCreate indicates that the ContainerSSH Docker module failed to create a container. This may be a
// temporary and retried or a permanent error message. Check the log message for details.
const EDockerFailedContainerCreate = "DOCKER_CONTAINER_CREATE_FAILED"

// MDockerContainerStart indicates that the ContainerSSH Docker module is starting the previously-created container.
const MDockerContainerStart = "DOCKER_CONTAINER_START"

// EDockerFailedContainerStart indicates that the ContainerSSH docker module failed to start the container. This message
// can either be temporary and retried or permanent. Check the log message for details.
const EDockerFailedContainerStart = "DOCKER_CONTAINER_START_FAILED"

// MDockerContainerStop indicates that the ContainerSSH Docker module is stopping the container.
const MDockerContainerStop = "DOCKER_CONTAINER_STOP"

// EDockerContainerStopFailed indicates that the ContainerSSH Docker module failed to stop the container. This message can
// be either temporary and retried or permanent. Check the log message for details.
const EDockerContainerStopFailed = "DOCKER_CONTAINER_STOP_FAILED"

// MDockerContainerRemove indicates that the ContainerSSH Docker module os removing the container.
const MDockerContainerRemove = "DOCKER_CONTAINER_REMOVE"

// EDockerFailedContainerRemove indicates that the ContainerSSH Docker module could not remove the container. This
// message may be temporary and retried or permanent. Check the log message for details.
const EDockerFailedContainerRemove = "DOCKER_CONTAINER_REMOVE_FAILED"

// MDockerContainerRemoveSuccessful indicates that the ContainerSSH Docker module has successfully removed the container.
const MDockerContainerRemoveSuccessful = "DOCKER_CONTAINER_REMOVE_SUCCESSFUL"

// MDockerContainerSignal indicates that the ContainerSSH Docker module is sending a signal to the container.
const MDockerContainerSignal = "DOCKER_CONTAINER_SIGNAL"

// EDockerFailedContainerSignal indicates that the ContainerSSH Docker module has failed to send a signal to the
// container.
const EDockerFailedContainerSignal = "DOCKER_CONTAINER_SIGNAL_FAILED"

// EDockerShuttingDown indicates that the ContainerSSH Docker module is shutting down a container.
const EDockerShuttingDown = "DOCKER_CONTAINER_SHUTTING_DOWN"

// MDockerExec indicates that the ContainerSSH Docker module is creating an execution. This may be in connection mode, or
// it may be the module internally using the exec mechanism to deliver a payload into the container.
const MDockerExec = "DOCKER_EXEC"

// MDockerExecAttach indicates that the ContainerSSH Docker module is attaching to the previously-created execution.
const MDockerExecAttach = "DOCKER_EXEC_ATTACH"

// EDockerFailedExecAttach indicates that the ContainerSSH Docker module could not attach to the previously-created
// execution.
const EDockerFailedExecAttach = "DOCKER_EXEC_ATTACH_FAILED"

// MDockerExecCreate indicates that the ContainerSSH Docker module is creating an execution.
const MDockerExecCreate = "DOCKER_EXEC_CREATE"

// EDockerFailedExecCreate indicates that the ContainerSSH Docker module has failed to create an execution. This can be
// temporary and retried or permanent. See the error message for details.
const EDockerFailedExecCreate = "DOCKER_EXEC_CREATE_FAILED"

// EDockerFailedPIDRead indicates that the ContainerSSH Docker module has failed to read the process ID from the
// ContainerSSH Guest Agent. This is most likely because the guest image does not contain the guest agent, but guest
// agent support has been enabled.
const EDockerFailedPIDRead = "DOCKER_EXEC_PID_READ_FAILED"

// MDockerResizing indicates that the ContainerSSH Docker module is resizing the console.
const MDockerResizing = "DOCKER_EXEC_RESIZE"

// EDockerFailedResize indicates that the ContainerSSH Docker module failed to resize the console.
const EDockerFailedResize = "DOCKER_EXEC_RESIZE_FAILED"

// MDockerExecSignal indicates that the ContainerSSH Docker module is delivering a signal in container mode.
const MDockerExecSignal = "DOCKER_EXEC_SIGNAL"

// EDockerFailedExecSignal indicates that the ContainerSSH Docker module failed to deliver a signal.
const EDockerFailedExecSignal = "DOCKER_EXEC_SIGNAL_FAILED"

// EDockerCannotSendSignalNoAgent indicates that the ContainerSSH Docker module failed to deliver a signal because
// ContainerSSH Guest Agent support is disabled.
const EDockerCannotSendSignalNoAgent = "DOCKER_EXEC_SIGNAL_FAILED_NO_AGENT"

// MDockerExecSignalSuccessful indicates that the ContainerSSH Docker module successfully delivered the requested
// signal.
const MDockerExecSignalSuccessful = "DOCKER_EXEC_SIGNAL_SUCCESSFUL"

// MDockerExitCode indicates that the ContainerSSH Docker module is fetching the exit code from the program.
const MDockerExitCode = "DOCKER_EXIT_CODE"

// EDockerContainerRestarting indicates that the ContainerSSH Docker module could not fetch the exit code from the
// program because the container is restarting. This is typically a misconfiguration as ContainerSSH containers should
// not automatically restart.
const EDockerContainerRestarting = "DOCKER_EXIT_CODE_CONTAINER_RESTARTING"

// EDockerFetchingExitCodeFailed indicates that the ContainerSSH Docker module has failed to fetch the exit code of the
// program.
const EDockerFetchingExitCodeFailed = "DOCKER_EXIT_CODE_FAILED"

// EDockerNegativeExitCode indicates that the ContainerSSH Docker module has received a negative exit code from Docker.
// This should never happen and is most likely a bug.
const EDockerNegativeExitCode = "DOCKER_EXIT_CODE_NEGATIVE"

// EDockerStillRunning indicates that the ContainerSSH Docker module could not fetch the program exit code because the
// program is still running. This error may be temporary and retried or permanent.
const EDockerStillRunning = "DOCKER_EXIT_CODE_STILL_RUNNING"

// MDockerImageList indicates that the ContainerSSH Docker module is listing the locally present container images to
// determine if the specified container image needs to be pulled.
const MDockerImageList = "DOCKER_IMAGE_LISTING"

// EDockerFailedImageList indicates that the ContainerSSH Docker module failed to list the images present in the local
// Docker daemon. This is used to determine if the image needs to be pulled. This can be because the Docker daemon is
// not reachable, the certificate is invalid, or there is something else interfering with listing the images.
const EDockerFailedImageList = "DOCKER_IMAGE_LISTING_FAILED"

// MDockerImagePull indicates that the ContainerSSH Docker module is pulling the container image.
const MDockerImagePull = "DOCKER_IMAGE_PULL"

// EDockerFailedImagePull indicates that the ContainerSSH Docker module failed to pull the specified container image.
// This can be because of connection issues to the Docker daemon, or because the Docker daemon itself can't pull the
// image. If you don't intend to have the image pulled you should set the `ImagePullPolicy` to `Never`. See the
// Docker documentation for details.
const EDockerFailedImagePull = "DOCKER_IMAGE_PULL_FAILED"

// MDockerImagePullNeeded indicates that the ContainerSSH Docker module is checking if an image pull is needed.
const MDockerImagePullNeeded = "DOCKER_IMAGE_PULL_NEEDED_CHECKING"

// EDockerProgramAlreadyRunning indicates that the ContainerSSH Docker module can't execute the request because the
// program is already running. This is a client error.
const EDockerProgramAlreadyRunning = "DOCKER_PROGRAM_ALREADY_RUNNING"

// EDockerFailedSignalNoPID indicates that the ContainerSSH Docker module can't deliver a signal because no PID has been
// recorded. This is most likely because guest agent support is disabled.
const EDockerFailedSignalNoPID = "DOCKER_SIGNAL_FAILED_NO_PID"

// EDockerFailedInputStream indicates that the ContainerSSH Docker module failed to stream stdin to the Docker engine.
const EDockerFailedInputStream = "DOCKER_STREAM_INPUT_FAILED"

// EDockerFailedOutputStream indicates that the ContainerSSH Docker module failed to stream stdout and stderr from the
// Docker engine.
const EDockerFailedOutputStream = "DOCKER_STREAM_OUTPUT_FAILED"

// EDockerSubsystemNotSupported indicates that the ContainerSSH Docker module is not configured to run the requested
// subsystem.
const EDockerSubsystemNotSupported = "DOCKER_SUBSYSTEM_NOT_SUPPORTED"

// EDockerGuestAgentDisabled indicates that the ContainerSSH Guest Agent has been disabled, which is strongly
// discouraged. ContainerSSH requires the guest agent to be installed in the container image to facilitate all SSH
// features. Disabling the guest agent will result in breaking the expectations a user has towards an SSH server. We
// provide the ability to disable guest agent support only for cases where the guest agent binary cannot be installed in
// the image at all.
const EDockerGuestAgentDisabled = "DOCKER_GUEST_AGENT_DISABLED"

// EDockerProgramNotRunning indicates that the user requested an action that can only be performed when
// a program is running, but there is currently no program running.
const EDockerProgramNotRunning = "DOCKER_PROGRAM_NOT_RUNNING"

// EDockerRunRemoved indicates that the configuration contained a dockerrun configuration segment, but this backend was
// removed since ContainerSSH 0.5. To fix this error please remove the dockerrun segment from your configuration or
// configuration server response. For details please see https://containerssh.io/deprecations/dockerrun/ .
const EDockerRunRemoved = "DOCKER_RUN_REMOVED"

// EDockerWriteFileFailed indicates that the ContainerSSH docker backend failed to write to a file in the container
const EDockerWriteFileFailed = "DOCKER_FILE_WRITE_FAILED"

// EDockerFileWrite indicates that the ContainerSSH docker backend wrote a file inside the container
const MDockerFileWrite = "DOCKER_FILE_WRITE"
