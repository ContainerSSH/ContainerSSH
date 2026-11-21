package metricsintegration

// MetricNameConnections is the number of connections since start.
const MetricNameConnections = "containerssh_ssh_connections_total"

// MetricHelpConnections is the help text for the number of connections since start.
const MetricHelpConnections = "Number of connections since start"

// MetricNameCurrentConnections is the number of currently open SSH connections.
const MetricNameCurrentConnections = "containerssh_ssh_current_connections"

// MetricHelpCurrentConnections is th ehelp text for the number of currently open SSH connections.
const MetricHelpCurrentConnections = "Current open SSH connections"

// MetricNameSuccessfulHandshake is the number of successful SSH handshakes since start.
const MetricNameSuccessfulHandshake = "containerssh_ssh_successful_handshakes_total"

// MetricHelpSuccessfulHandshake is the help text for the number of successful SSH handshakes since start.
const MetricHelpSuccessfulHandshake = "Successful SSH handshakes since start"

// MetricNameFailedHandshake is the number of failed SSH handshakes since start.
const MetricNameFailedHandshake = "containerssh_ssh_failed_handshakes_total"

// MetricHelpFailedHandshake is the help text for the number of failed SSH handshakes since start.
const MetricHelpFailedHandshake = "Failed SSH handshakes since start"
