package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
)

// LogConfig describes the logging settings.
type LogConfig struct {
	// Level describes the minimum level to log at
	Level LogLevel `json:"level" yaml:"level" default:"5"`

	// Format describes the log message format
	Format LogFormat `json:"format" yaml:"format" default:"ljson"`

	// Destination is the target to write the log messages to.
	Destination LogDestination `json:"destination" yaml:"destination" default:"stdout"`

	// File is the log file to write to if Destination is set to "file".
	File string `json:"file" yaml:"file" default:"/var/log/containerssh/containerssh.log"`

	// Syslog configures the syslog destination.
	Syslog SyslogConfig `json:"syslog" yaml:"syslog"`

	// T is the Go test for logging purposes.
	T *testing.T `json:"-" yaml:"-"`

	// Stdout is the standard output used by the LogDestinationStdout destination.
	Stdout io.Writer `json:"-" yaml:"-"`
}

// Validate validates the log configuration.
func (c *LogConfig) Validate() error {
	if err := c.Level.Validate(); err != nil {
		return wrap(err, "level")
	}
	if err := c.Format.Validate(); err != nil {
		return wrap(err, "format")
	}
	if err := c.Destination.Validate(); err != nil {
		return wrap(err, "destination")
	}
	if c.Destination == LogDestinationTest && c.T == nil {
		return fmt.Errorf("test log destination selected but no test case provided")
	}
	return nil
}

// region LogLevel

// LogLevel syslog-style log level identifiers
type LogLevel int8

// Supported values for LogLevel
const (
	LogLevelDebug     LogLevel = 7
	LogLevelInfo      LogLevel = 6
	LogLevelNotice    LogLevel = 5
	LogLevelWarning   LogLevel = 4
	LogLevelError     LogLevel = 3
	LogLevelCritical  LogLevel = 2
	LogLevelAlert     LogLevel = 1
	LogLevelEmergency LogLevel = 0
)

// UnmarshalJSON decodes a JSON level string to a level type.
func (level *LogLevel) UnmarshalJSON(data []byte) error {
	var levelString LogLevelString
	if err := json.Unmarshal(data, &levelString); err != nil {
		unmarshalError := &json.UnmarshalTypeError{}
		if errors.As(err, &unmarshalError) {
			type levelAlias LogLevel
			var l levelAlias
			if err = json.Unmarshal(data, &l); err != nil {
				return err
			}
			*level = LogLevel(l)
		}
		return err
	}
	l, err := levelString.ToLevel()
	if err != nil {
		return err
	}
	*level = l
	return nil
}

// MarshalJSON marshals a level number to a JSON string
func (level LogLevel) MarshalJSON() ([]byte, error) {
	levelString, err := level.Name()
	if err != nil {
		return nil, err
	}
	return json.Marshal(levelString)
}

// UnmarshalYAML decodes a YAML level string to a level type.
func (level *LogLevel) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var levelString LogLevelString
	if err := unmarshal(&levelString); err != nil {
		return err
	}
	l, err := levelString.ToLevel()
	if err != nil {
		type levelAlias LogLevel
		var l2 levelAlias
		if err2 := unmarshal(&l2); err2 != nil {
			return err
		}
		*level = LogLevel(l2)
		return nil
	}
	*level = l
	return nil
}

// MarshalYAML creates a YAML text representation from a numeric level
func (level LogLevel) MarshalYAML() (interface{}, error) {
	return level.Name()
}

// Name Convert the int level to the string representation
func (level *LogLevel) Name() (LogLevelString, error) {
	switch *level {
	case LogLevelDebug:
		return LevelDebugString, nil
	case LogLevelInfo:
		return LevelInfoString, nil
	case LogLevelNotice:
		return LevelNoticeString, nil
	case LogLevelWarning:
		return LevelWarningString, nil
	case LogLevelError:
		return LevelErrorString, nil
	case LogLevelCritical:
		return LevelCriticalString, nil
	case LogLevelAlert:
		return LevelAlertString, nil
	case LogLevelEmergency:
		return LevelEmergencyString, nil
	}
	return "", fmt.Errorf("invalid log level (%d)", level)
}

// MustName Convert the int level to the string representation and panics if the name is not valid.
func (level LogLevel) MustName() LogLevelString {
	name, err := level.Name()
	if err != nil {
		panic(err)
	}
	return name
}

// String Convert the int level to the string representation. Panics if the level is not valid.
func (level LogLevel) String() string {
	return string(level.MustName())
}

// Validate if the log level has a valid value
func (level LogLevel) Validate() error {
	if level < LogLevelEmergency || level > LogLevelDebug {
		return fmt.Errorf("invalid log level (%d)", level)
	}
	return nil
}

// endregion

// region LogLevelString

// LogLevelString is a type for supported log level strings
type LogLevelString string

// List of valid string values for log levels
const (
	LevelDebugString     LogLevelString = "debug"
	LevelInfoString      LogLevelString = "info"
	LevelNoticeString    LogLevelString = "notice"
	LevelWarningString   LogLevelString = "warning"
	LevelErrorString     LogLevelString = "error"
	LevelCriticalString  LogLevelString = "crit"
	LevelAlertString     LogLevelString = "alert"
	LevelEmergencyString LogLevelString = "emerg"
)

// ToLevel convert the string level to the int representation
func (level LogLevelString) ToLevel() (LogLevel, error) {
	switch level {
	case LevelDebugString:
		return LogLevelDebug, nil
	case LevelInfoString:
		return LogLevelInfo, nil
	case LevelNoticeString:
		return LogLevelNotice, nil
	case LevelWarningString:
		return LogLevelWarning, nil
	case LevelErrorString:
		return LogLevelError, nil
	case LevelCriticalString:
		return LogLevelCritical, nil
	case LevelAlertString:
		return LogLevelAlert, nil
	case LevelEmergencyString:
		return LogLevelEmergency, nil
	}
	return -1, fmt.Errorf("invalid log level (%s)", level)
}

// endregion

// region LogFormat

// LogFormat is the logging format to use.
//
//swagger:enum
type LogFormat string

const (
	// LogFormatLJSON is a newline-delimited JSON log format.
	LogFormatLJSON LogFormat = "ljson"
	// LogFormatText prints the logs as plain text.
	LogFormatText LogFormat = "text"
)

// Validate returns an error if the format is invalid.
func (format LogFormat) Validate() error {
	switch format {
	case LogFormatLJSON:
	case LogFormatText:
	default:
		return fmt.Errorf("invalid log format: %s", format)
	}
	return nil
}

// endregion

// region LogDestination

// LogDestination is the output to write to.
//
//swagger:enum
type LogDestination string

const (
	// LogDestinationStdout is writing log messages to the standard output.
	LogDestinationStdout LogDestination = "stdout"
	// LogDestinationFile is writing the log messages to a file.
	LogDestinationFile LogDestination = "file"
	// LogDestinationSyslog is writing log messages to syslog.
	LogDestinationSyslog LogDestination = "syslog"
	// LogDestinationTest writes the logs to the *testing.T facility.
	LogDestinationTest LogDestination = "test"
)

// Validate validates the output target.
func (o LogDestination) Validate() error {
	switch o {
	case LogDestinationStdout:
	case LogDestinationFile:
	case LogDestinationSyslog:
	case LogDestinationTest:
	default:
		return fmt.Errorf("invalid destination: %s", o)
	}
	return nil
}

// endregion

// region Syslog

// LogFacility describes the syslog facility log messages are sent to.
type LogFacility int

const (
	// LogFacilityKern are kernel messages.
	LogFacilityKern LogFacility = 0
	// LogFacilityUser are user level messages.
	LogFacilityUser LogFacility = 1
	// LogFacilityMail are user mail log messages.
	LogFacilityMail LogFacility = 2
	// LogFacilityDaemon are daemon messages.
	LogFacilityDaemon LogFacility = 3
	// LogFacilityAuth are authentication messages.
	LogFacilityAuth LogFacility = 4
	// LogFacilitySyslog are syslog-specific messages.
	LogFacilitySyslog LogFacility = 5
	// LogFacilityLPR are printer messages.
	LogFacilityLPR LogFacility = 6
	// LogFacilityNews are news messages.
	LogFacilityNews LogFacility = 7
	// LogFacilityUUCP are UUCP subsystem messages.
	LogFacilityUUCP LogFacility = 8
	// LogFacilityCron are clock daemon messages.
	LogFacilityCron LogFacility = 9
	// LogFacilityAuthPriv are security/authorization messages.
	LogFacilityAuthPriv LogFacility = 10
	// LogFacilityFTP are FTP daemon messages.
	LogFacilityFTP LogFacility = 11
	// LogFacilityNTP are network time daemon messages.
	LogFacilityNTP LogFacility = 12
	// LogFacilityLogAudit are log audit messages.
	LogFacilityLogAudit LogFacility = 13
	// LogFacilityLogAlert are log alert messages.
	LogFacilityLogAlert LogFacility = 14
	// LogFacilityClock are clock daemon messages.
	LogFacilityClock LogFacility = 15

	// LogFacilityLocal0 are locally administered messages.
	LogFacilityLocal0 LogFacility = 16
	// LogFacilityLocal1 are locally administered messages.
	LogFacilityLocal1 LogFacility = 17
	// LogFacilityLocal2 are locally administered messages.
	LogFacilityLocal2 LogFacility = 18
	// LogFacilityLocal3 are locally administered messages.
	LogFacilityLocal3 LogFacility = 19
	// LogFacilityLocal4 are locally administered messages.
	LogFacilityLocal4 LogFacility = 20
	// LogFacilityLocal5 are locally administered messages.
	LogFacilityLocal5 LogFacility = 21
	// LogFacilityLocal6 are locally administered messages.
	LogFacilityLocal6 LogFacility = 22
	// LogFacilityLocal7 are locally administered messages.
	LogFacilityLocal7 LogFacility = 23
)

// Validate checks if the facility is valid.
func (f LogFacility) Validate() error {
	if _, ok := facilityToName[f]; !ok {
		return fmt.Errorf("invalid facility: %d", f)
	}
	return nil
}

// Name returns the facility name.
func (f LogFacility) Name() (LogFacilityString, error) {
	if name, ok := facilityToName[f]; ok {
		return name, nil
	}
	return "", fmt.Errorf("invalid facility: %d", f)
}

// MustName is identical to Name but panics if the facility is invalid.
func (f LogFacility) MustName() LogFacilityString {
	name, err := f.Name()
	if err != nil {
		panic(err)
	}
	return name
}

// LogFacilityString are facility names.
type LogFacilityString string

const (
	// LogFacilityStringKern are kernel messages.
	LogFacilityStringKern LogFacilityString = "kern"
	// LogFacilityStringUser are user level messages.
	LogFacilityStringUser LogFacilityString = "user"
	// LogFacilityStringMail are user mail log messages.
	LogFacilityStringMail LogFacilityString = "mail"
	// LogFacilityStringDaemon are daemon messages.
	LogFacilityStringDaemon LogFacilityString = "daemon"
	// LogFacilityStringAuth are authentication messages.
	LogFacilityStringAuth LogFacilityString = "auth"
	// LogFacilityStringSyslog are syslog-specific messages.
	LogFacilityStringSyslog LogFacilityString = "syslog"
	// LogFacilityStringLPR are printer messages.
	LogFacilityStringLPR LogFacilityString = "lpr"
	// LogFacilityStringNews are news messages.
	LogFacilityStringNews LogFacilityString = "news"
	// LogFacilityStringUUCP are UUCP subsystem messages.
	LogFacilityStringUUCP LogFacilityString = "uucp"
	// LogFacilityStringCron are clock daemon messages.
	LogFacilityStringCron LogFacilityString = "cron"
	// LogFacilityStringAuthPriv are security/authorization messages.
	LogFacilityStringAuthPriv LogFacilityString = "authpriv"
	// LogFacilityStringFTP are FTP daemon messages.
	LogFacilityStringFTP LogFacilityString = "ftp"
	// LogFacilityStringNTP are network time daemon messages.
	LogFacilityStringNTP LogFacilityString = "ntp"
	// LogFacilityStringLogAudit are log audit messages.
	LogFacilityStringLogAudit LogFacilityString = "logaudit"
	// LogFacilityStringLogAlert are log alert messages.
	LogFacilityStringLogAlert LogFacilityString = "logalert"
	// LogFacilityStringClock are clock daemon messages.
	LogFacilityStringClock LogFacilityString = "clock"

	// LogFacilityStringLocal0 are locally administered messages.
	LogFacilityStringLocal0 LogFacilityString = "local0"
	// LogFacilityStringLocal1 are locally administered messages.
	LogFacilityStringLocal1 LogFacilityString = "local1"
	// LogFacilityStringLocal2 are locally administered messages.
	LogFacilityStringLocal2 LogFacilityString = "local2"
	// LogFacilityStringLocal3 are locally administered messages.
	LogFacilityStringLocal3 LogFacilityString = "local3"
	// LogFacilityStringLocal4 are locally administered messages.
	LogFacilityStringLocal4 LogFacilityString = "local4"
	// LogFacilityStringLocal5 are locally administered messages.
	LogFacilityStringLocal5 LogFacilityString = "local5"
	// LogFacilityStringLocal6 are locally administered messages.
	LogFacilityStringLocal6 LogFacilityString = "local6"
	// LogFacilityStringLocal7 are locally administered messages.
	LogFacilityStringLocal7 LogFacilityString = "local7"
)

// Validate validates the facility string.
func (s LogFacilityString) Validate() error {
	if _, ok := nameToFacility[s]; !ok {
		return fmt.Errorf("invalid facility: %s", s)
	}
	return nil
}

// Number returns the facility number.
func (s LogFacilityString) Number() (LogFacility, error) {
	if val, ok := nameToFacility[s]; ok {
		return val, nil
	}
	return LogFacility(-1), fmt.Errorf("invalid facility: %s", s)
}

// MustNumber is identical to Number but panics instead of returning errors
func (s LogFacilityString) MustNumber() LogFacility {
	n, err := s.Number()
	if err != nil {
		panic(err)
	}
	return n
}

var facilityToName = map[LogFacility]LogFacilityString{
	LogFacilityKern:     LogFacilityStringKern,
	LogFacilityUser:     LogFacilityStringUser,
	LogFacilityMail:     LogFacilityStringMail,
	LogFacilityDaemon:   LogFacilityStringDaemon,
	LogFacilityAuth:     LogFacilityStringAuth,
	LogFacilitySyslog:   LogFacilityStringSyslog,
	LogFacilityLPR:      LogFacilityStringLPR,
	LogFacilityNews:     LogFacilityStringNews,
	LogFacilityUUCP:     LogFacilityStringUUCP,
	LogFacilityCron:     LogFacilityStringCron,
	LogFacilityAuthPriv: LogFacilityStringAuthPriv,
	LogFacilityFTP:      LogFacilityStringFTP,
	LogFacilityNTP:      LogFacilityStringNTP,
	LogFacilityLogAudit: LogFacilityStringLogAudit,
	LogFacilityLogAlert: LogFacilityStringLogAlert,
	LogFacilityClock:    LogFacilityStringClock,

	LogFacilityLocal0: LogFacilityStringLocal0,
	LogFacilityLocal1: LogFacilityStringLocal1,
	LogFacilityLocal2: LogFacilityStringLocal2,
	LogFacilityLocal3: LogFacilityStringLocal3,
	LogFacilityLocal4: LogFacilityStringLocal4,
	LogFacilityLocal5: LogFacilityStringLocal5,
	LogFacilityLocal6: LogFacilityStringLocal6,
	LogFacilityLocal7: LogFacilityStringLocal7,
}

var nameToFacility = map[LogFacilityString]LogFacility{
	LogFacilityStringKern:     LogFacilityKern,
	LogFacilityStringUser:     LogFacilityUser,
	LogFacilityStringMail:     LogFacilityMail,
	LogFacilityStringDaemon:   LogFacilityDaemon,
	LogFacilityStringAuth:     LogFacilityAuth,
	LogFacilityStringSyslog:   LogFacilitySyslog,
	LogFacilityStringLPR:      LogFacilityLPR,
	LogFacilityStringNews:     LogFacilityNews,
	LogFacilityStringUUCP:     LogFacilityUUCP,
	LogFacilityStringCron:     LogFacilityCron,
	LogFacilityStringAuthPriv: LogFacilityAuthPriv,
	LogFacilityStringFTP:      LogFacilityFTP,
	LogFacilityStringNTP:      LogFacilityNTP,
	LogFacilityStringLogAudit: LogFacilityLogAudit,
	LogFacilityStringLogAlert: LogFacilityLogAlert,
	LogFacilityStringClock:    LogFacilityClock,

	LogFacilityStringLocal0: LogFacilityLocal0,
	LogFacilityStringLocal1: LogFacilityLocal1,
	LogFacilityStringLocal2: LogFacilityLocal2,
	LogFacilityStringLocal3: LogFacilityLocal3,
	LogFacilityStringLocal4: LogFacilityLocal4,
	LogFacilityStringLocal5: LogFacilityLocal5,
	LogFacilityStringLocal6: LogFacilityLocal6,
	LogFacilityStringLocal7: LogFacilityLocal7,
}

// SyslogConfig is the configuration for syslog logging.
//
//goland:noinspection GoVetStructTag
type SyslogConfig struct {
	// Destination is the socket to send logs to. Can be a local path to unix sockets as well as UDP destinations.
	Destination string `json:"destination" yaml:"destination" default:"/dev/log"`
	// Facility logs to the specified syslog facility.
	Facility LogFacilityString `json:"facility" yaml:"facility" default:"auth"`
	// Tag is the syslog tag to log with.
	Tag string `json:"tag" yaml:"tag" default:"ContainerSSH"`
	// Pid is a setting to append the current process ID to the tag.
	Pid bool `json:"pid" yaml:"pid" default:"false"`

	// Connection is the connection to the Syslog server. This is only present after Validate has been called.
	Connection net.Conn `json:"-" yaml:"-"`
}

// Validate validates the syslog configuration
func (c *SyslogConfig) Validate() error {
	destination := "/dev/log"
	if c.Destination != "" {
		destination = c.Destination
	}
	if strings.HasPrefix(c.Destination, "/") {
		connection, err := net.Dial("unix", destination)
		if err != nil {
			connection, err = net.Dial("unixgram", destination)
			if err != nil {
				return fmt.Errorf("failed to open UNIX socket to %s (%w)", c.Destination, err)
			}
		}
		c.Connection = connection
	} else {
		connection, err := net.Dial("udp", destination)
		if err != nil {
			return fmt.Errorf("failed to open UDP socket to %s (%w)", c.Destination, err)
		}
		c.Connection = connection
	}
	if err := c.Facility.Validate(); err != nil {
		return err
	}
	return nil
}

// endregion
