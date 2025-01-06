package message

// ELogWriteFailed indicates that ContainerSSH cannot write to the specified log file. This usually happens because the
// underlying filesystem is full or the log is located on a non-local storage (e.g. NFS), which is not supported.
const ELogWriteFailed = "LOG_WRITE_FAILED"

// ELogRotateFailed indicates that ContainerSSH cannot rotate the logs as requested because of an underlying error.
const ELogRotateFailed = "LOG_ROTATE_FAILED"

// ELogFileOpenFailed indicates that ContainerSSH failed to open the specified log file.
const ELogFileOpenFailed = "LOG_FILE_OPEN_FAILED"

// EUnknownError indicates that the error cannot be identified. If you see this in a log that is a bug and should be reported.
const EUnknownError = "UNKNOWN_ERROR"

// MTest indicates that this message should only be seen in unit and component tests, never in production.
const MTest = "TEST"
