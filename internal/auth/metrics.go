package auth

const (
	// MetricNameAuthBackendRequests is the number of requests to the config server.
	MetricNameAuthBackendRequests = "containerssh_auth_server_requests_total"

	// MetricNameAuthBackendFailure is the number of request failures to the configuration backend.
	MetricNameAuthBackendFailure = "containerssh_auth_server_failures_total"

	// MetricNameAuthSuccess captures the number of successful authentication attempts.
	MetricNameAuthSuccess = "containerssh_auth_success_total"

	// MetricNameAuthFailure captures the number of failed authentication attempts.
	MetricNameAuthFailure = "containerssh_auth_failures_total"
)
