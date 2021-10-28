package message

//go:generate go run github.com/containerssh/libcontainerssh/cmd/generate-message-codes auditlog.go AUDITLOG.md "Audit log"

// MAuditLogMultipartUpload indicates that ContainerSSH is starting a new S3 multipart upload.
const MAuditLogMultipartUpload = "AUDIT_S3_MULTIPART_UPLOAD"

// EAuditLogMultipartUploadInitializationFailed indicates that ContainerSSH failed to initialize a new multipart upload
// to the S3-compatible object storage. Check if the S3 configuration is correct and the provided S3 access key and
// secrets have permissions to start a multipart upload.
const EAuditLogMultipartUploadInitializationFailed = "AUDIT_S3_MULTIPART_UPLOAD_INITIALIZATION_FAILED"

// MAuditLogMultipartPartUploading indicates that ContainerSSH is uploading a part of an audit log to the S3-compatible
// object storage.
const MAuditLogMultipartPartUploading = "AUDIT_S3_MULTIPART_PART_UPLOADING"

// EAuditLogMultipartPartUploadFailed indicates that ContainerSSH failed to upload a part to the S3-compatible object
// storage. Check the message for details.
const EAuditLogMultipartPartUploadFailed = "AUDIT_S3_MULTIPART_PART_UPLOAD_FAILED"

// MAuditLogMultipartPartUploadComplete indicates that ContainerSSH completed the upload of an audit log part to the
// S3-compatible object storage.
const MAuditLogMultipartPartUploadComplete = "AUDIT_S3_MULTIPART_PART_UPLOAD_COMPLETE"

// MAuditLogMultipartUploadFinalizing indicates that ContainerSSH has uploaded all audit log parts and is now finalizing
// the multipart upload.
const MAuditLogMultipartUploadFinalizing = "AUDIT_S3_MULTIPART_UPLOAD_FINALIZING"

// EAuditLogMultipartUploadFinalizationFailed indicates that ContainerSSH has uploaded all audit log parts, but could
// not finalize the multipart upload.
const EAuditLogMultipartUploadFinalizationFailed = "AUDIT_S3_MULTIPART_UPLOAD_FINALIZATION_FAILED"

// EAuditLogMultipartUploadFinalized indicates that ContainerSSH has uploaded all audit log parts and has successfully
// finalized the upload.
const EAuditLogMultipartUploadFinalized = "AUDIT_S3_MULTIPART_UPLOAD_FINALIZED"

// EAuditLogMultipartFailedAbort indicates that ContainerSSH failed aborting a multipart upload from a previously
// crashed ContainerSSH run.
const EAuditLogMultipartFailedAbort = "AUDIT_S3_MULTIPART_FAILED_ABORT"

// EAuditLogMultipartFailedList indicates that ContainerSSH failed to list multipart uploads on the object storage
// bucket. This is needed to abort uploads from a
// previously crashed ContainerSSH run.
const EAuditLogMultipartFailedList = "AUDIT_S3_MULTIPART_FAILED_LIST"

// MAuditLogSingleUpload indicates that ContainerSSH is uploading the full audit log in a single upload to the
// S3-compatible object storage. This happens when the audit log size is below the minimum size for a multi-part upload.
const MAuditLogSingleUpload = "AUDIT_S3_SINGLE_UPLOAD"

// MAuditLogSingleUploadFailed indicates that ContainerSSH failed to upload the audit log as a single upload.
const MAuditLogSingleUploadFailed = "AUDIT_S3_SINGLE_UPLOAD_FAILED"

// MAuditLogSingleUploadComplete indicates that ContainerSSH successfully uploaded the audit log as a single upload.
const MAuditLogSingleUploadComplete = "AUDIT_S3_SINGLE_UPLOAD_COMPLETE"

// EAuditLogFailedCreatingMetadataFile indicates that ContainerSSH failed to create the metadata file for the S3 upload
// in the local temporary directory. Check if the local directory specified is writable and has enough disk space.
const EAuditLogFailedCreatingMetadataFile = "AUDIT_S3_FAILED_CREATING_METADATA_FILE"

// EAuditLogFailedReadingMetadataFile indicates that ContainerSSH failed to read the metadata file for the S3 upload in
// the local temporary directory. Check if the local directory specified is readable and the files have not been corrupted.
const EAuditLogFailedReadingMetadataFile = "AUDIT_S3_FAILED_READING_METADATA_FILE"

// EAuditLogCannotCloseMetadataFileHandle indicates that ContainerSSH could not close the metadata file in the local
// folder. This typically happens when the local folder is on an NFS share. (Running an audit log on an NFS share is
// NOT supported.)
const EAuditLogCannotCloseMetadataFileHandle = "AUDIT_S3_CANNOT_CLOSE_METADATA_FILE_HANDLE"

// EAuditLogFailedMetadataJSONEncoding indicates that ContainerSSH failed to encode the metadata file. This is a bug,
// please report it.
const EAuditLogFailedMetadataJSONEncoding = "AUDIT_S3_FAILED_METADATA_JSON_ENCODING"

// EAuditLogFailedWritingMetadataFile indicates that ContainerSSH failed to write the local metadata file. Please check
// if your disk has enough disk space.
const EAuditLogFailedWritingMetadataFile = "AUDIT_S3_FAILED_WRITING_METADATA_FILE"

// EAuditLogFailedQueueStat indicates that ContainerSSH failed to stat the queue file. This usually happens when the
// local directory is being manually manipulated.
const EAuditLogFailedQueueStat = "AUDIT_S3_FAILED_STAT_QUEUE_ENTRY"

// EAuditLogNoSuchQueueEntry indicates that ContainerSSH was trying to upload an audit log from the metadata file, but
// the audit log does not exist.
const EAuditLogNoSuchQueueEntry = "AUDIT_S3_NO_SUCH_QUEUE_ENTRY"

// EAuditLogRemoveAuditLogFailed indicates that ContainerSSH failed to remove an uploaded audit log from the local
// directory. This usually happens on Windows when a different process has the audit log open. (This is not a supported
// setup.)
const EAuditLogRemoveAuditLogFailed = "AUDIT_S3_REMOVE_FAILED"

// EAuditLogCloseAuditLogFileFailed indicates that ContainerSSH failed to close an audit log file in the local
// directory. This usually happens when the local directory is on an NFS share. (This is NOT supported.)
const EAuditLogCloseAuditLogFileFailed = "AUDIT_S3_CLOSE_FAILED"

// EAuditLogMultipartAborting indicates that ContainerSSH is aborting a multipart upload. Check the log message for
// details.
const EAuditLogMultipartAborting = "AUDIT_S3_MULTIPART_ABORTING"

// MAuditLogRecovering indicates that ContainerSSH found a previously aborted multipart upload locally and is now
// attempting to recover the upload.
const MAuditLogRecovering = "AUDIT_S3_RECOVERING"

// EAuditLogStorageCloseFailed indicates that ContainerSSH failed to close the audit log storage handler.
const EAuditLogStorageCloseFailed = "AUDIT_STORAGE_CLOSE_FAILED"

// EAuditLogStorageNotReadable indicates that The configured storage cannot be read from.
const EAuditLogStorageNotReadable = "AUDIT_STORAGE_NOT_READABLE"
