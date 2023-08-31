package message

// EKubernetesFailedOutputCloseWriting indicates that the ContainerSSH Kubernetes module attempted to close the output
// (stdout and stderr) for writing but failed to do so.
const EKubernetesFailedOutputCloseWriting = "KUBERNETES_CLOSE_OUTPUT_FAILED"

// MKubernetesPIDReceived indicates that the ContainerSSH Kubernetes module has received a PID from the Kubernetes guest
// agent.
const MKubernetesPIDReceived = "KUBERNETES_PID_RECEIVED"

// EKubernetesConfigError indicates that the ContainerSSH Kubernetes module detected a configuration error. Please check
// your configuration.
const EKubernetesConfigError = "KUBERNETES_CONFIG_ERROR"

// MKubernetesPodAttach indicates that the ContainerSSH Kubernetes module is attaching to a pod in session mode.
const MKubernetesPodAttach = "KUBERNETES_POD_ATTACH"

// MKubernetesPodCreate indicates that the ContainerSSH Kubernetes module is creating a pod.
const MKubernetesPodCreate = "KUBERNETES_POD_CREATE"

// MKubernetesPodWait indicates that the ContainerSSH Kubernetes module is waiting for the pod to come up.
const MKubernetesPodWait = "KUBERNETES_POD_WAIT"

// MKubernetesUsernameTooLong indicates that the users username is too long to be provided as a label in the k8s pod.
// The containerssh_username label is unavailable on that users pod.
const MKubernetesUsernameTooLong = "KUBERNETES_USERNAME_TOO_LONG"

// MKubernetesPodWaitFailed indicates that the ContainerSSH Kubernetes module failed to wait for the pod to come up.
// Check the error message for details.
const MKubernetesPodWaitFailed = "KUBERNETES_POD_WAIT_FAILED"

// EKubernetesFailedPodCreate indicates that the ContainerSSH Kubernetes module failed to create a pod. This may be a
// temporary and retried or a permanent error message. Check the log message for details.
const EKubernetesFailedPodCreate = "KUBERNETES_POD_CREATE_FAILED"

// MKubernetesPodRemove indicates that the ContainerSSH Kubernetes module is removing a pod.
const MKubernetesPodRemove = "KUBERNETES_POD_REMOVE"

// EKubernetesFailedPodRemove indicates that the ContainerSSH Kubernetes module could not remove the pod. This message
// may be temporary and retried or permanent. Check the log message for details.
const EKubernetesFailedPodRemove = "KUBERNETES_POD_REMOVE_FAILED"

// MKubernetesPodRemoveSuccessful indicates that the ContainerSSH Kubernetes module has successfully removed the pod.
const MKubernetesPodRemoveSuccessful = "KUBERNETES_POD_REMOVE_SUCCESSFUL"

// EKubernetesShuttingDown indicates that the ContainerSSH Kubernetes module is shutting down a pod.
const EKubernetesShuttingDown = "KUBERNETES_POD_SHUTTING_DOWN"

// MKubernetesExec indicates that the ContainerSSH Kubernetes module is creating an execution. This may be in connection
// mode, or it may be the module internally using the exec mechanism to deliver a payload into the pod.
const MKubernetesExec = "KUBERNETES_EXEC"

// MKubernetesResizing indicates that the ContainerSSH Kubernetes module is resizing the terminal window.
const MKubernetesResizing = "KUBERNETES_EXEC_RESIZE"

// EKubernetesFailedResize indicates that the ContainerSSH Kubernetes module failed to resize the console.
const EKubernetesFailedResize = "KUBERNETES_EXEC_RESIZE_FAILED"

// MKubernetesExecSignal indicates that the ContainerSSH Kubernetes module is delivering a signal.
const MKubernetesExecSignal = "KUBERNETES_EXEC_SIGNAL"

// EKubernetesFailedExecSignal indicates that the ContainerSSH Kubernetes module failed to deliver a signal.
const EKubernetesFailedExecSignal = "KUBERNETES_EXEC_SIGNAL_FAILED"

// EKubernetesCannotSendSignalNoAgent indicates that the ContainerSSH Kubernetes module failed to deliver a signal because guest
// agent support is disabled.
const EKubernetesCannotSendSignalNoAgent = "KUBERNETES_EXEC_SIGNAL_FAILED_NO_AGENT"

// MKubernetesExecSignalSuccessful indicates that the ContainerSSH Kubernetes module successfully delivered the requested signal.
const MKubernetesExecSignalSuccessful = "KUBERNETES_EXEC_SIGNAL_SUCCESSFUL"

// MKubernetesFileModification indicates that the ContainerSSH Kubernetes module is modifying a file on the container based on connection metadata
const MKubernetesFileModification = "KUBERNETES_FILE_WRITE"

// EKubernetesFileModificationFailed indicates that the ContainerSSH Kubernetes module failed to modify a file on the container
const EKubernetesFileModificationFailed = "KUBERNETES_FILE_WRITE_FAILED"

// EKubernetesFetchingExitCodeFailed indicates that the ContainerSSH Kubernetes module has failed to fetch the exit code of the
// program.
const EKubernetesFetchingExitCodeFailed = "KUBERNETES_EXIT_CODE_FAILED"

// EKubernetesProgramAlreadyRunning indicates that the ContainerSSH Kubernetes module can't execute the request because
// the program is already running. This is a client error.
const EKubernetesProgramAlreadyRunning = "KUBERNETES_PROGRAM_ALREADY_RUNNING"

// EKubernetesFailedSignalNoPID indicates that the ContainerSSH Kubernetes module can't deliver a signal because no PID
// has been recorded. This is most likely because guest agent support is disabled.
const EKubernetesFailedSignalNoPID = "KUBERNETES_SIGNAL_FAILED_NO_PID"

// EKubernetesFailedSignalExited indicates that the ContainerSSH Kubernetes module can't deliver a signal because the
// program already exited.
const EKubernetesFailedSignalExited = "KUBERNETES_SIGNAL_FAILED_EXITED"

// EKubernetesSubsystemNotSupported indicates that the ContainerSSH Kubernetes module is not configured to run the
// requested subsystem.
const EKubernetesSubsystemNotSupported = "KUBERNETES_SUBSYSTEM_NOT_SUPPORTED"

// EKubernetesGuestAgentDisabled indicates that the ContainerSSH Guest Agent has been disabled, which is strongly
// discouraged. ContainerSSH requires the guest agent to be installed in the pod image to facilitate all SSH features.
// Disabling the guest agent will result in breaking the expectations a user has towards an SSH server. We provide the
// ability to disable guest agent support only for cases where the guest agent binary cannot be installed in the image
// at all.
const EKubernetesGuestAgentDisabled = "KUBERNETES_GUEST_AGENT_DISABLED"

// EKubernetesProgramNotRunning indicates that the user requested an action that can only be performed when
// a program is running, but there is currently no program running.
const EKubernetesProgramNotRunning = "KUBERNETES_PROGRAM_NOT_RUNNING"

// EKubeRunRemoved indicates that the configuration contained a kuberun configuration segment, but this backend was
// removed since ContainerSSH 0.5. To fix this error please remove the kuberun segment from your configuration or
// configuration server response. For details please see https://containerssh.io/deprecations/kuberun/ .
const EKubeRunRemoved = "KUBERNETES_KUBERUN_REMOVED"

// EKubernetesForwardingFailed indicates that the ContainerSSH Kubernetes backend failed to initialize the forwarding backend
const EKubernetesForwardingFailed = "KUBERNETES_FORWARDING_FAILED"

// EKubernetesAgentFailed indicates that the ContainerSSH Kubernetes backend failed to start the ContainerSSH agent or the agent exited unexpectedly
const EKubernetesAgentFailed = "KUBERNETES_AGENT_FAILED"

// MKubernetesAgentLog indicates a log message from the ContainerSSH agent running within a user container.
// Note that the agent is normally run with the users credentials and as such all log output is to be considered UNTRUSTED and should only be used for debugging purposes
const MKubernetesAgentLog = "KUBERNETES_AGENT_LOG"
