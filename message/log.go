package message

// ContainerSSH cannot write to the specified log file. This usually happens because the underlying filesystem is full
// or the log is located on a non-local storage (e.g. NFS), which is not supported.
const ELogWriteFailed = "LOG_WRITE_FAILED"

// ContainerSSH cannot rotate the logs as requested because of an underlying error.
const ELogRotateFailed = "LOG_ROTATE_FAILED"

// ContainerSSH failed to open the specified log file.
const ELogFileOpenFailed = "LOG_FILE_OPEN_FAILED"

// This is an untyped error. If you see this in a log that is a bug and should be reported.
const EUnknownError = "UNKNOWN_ERROR"

// This is message that should only be seen in unit and component tests, never in production.
//goland:noinspection GoUnusedConst
const MTest = "TEST"
