package config

import (
	"errors"
	"fmt"
	"strings"
)

func newError(option string, message string, args ...interface{}) error {
	return &configError{
		message:    fmt.Sprintf(message, args...),
		optionPath: []string{option},
	}
}

func wrapWithMessage(err error, option string, message string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	var typedErr *configError
	if errors.As(err, &typedErr) {
		if message == "" {
			message = typedErr.message
		} else {
			message = fmt.Sprintf(message, args...)
		}
		return &configError{
			message:    message,
			optionPath: append([]string{option}, typedErr.optionPath...),
			cause:      err,
		}
	} else {
		if message == "" {
			message = err.Error()
		} else {
			message = fmt.Sprintf(message, args...) + " (" + err.Error() + ")"
		}
		return &configError{
			message:    message,
			optionPath: []string{option},
			cause:      err,
		}
	}
}

func wrap(err error, option string) error {
	return wrapWithMessage(err, option, "")
}

type configError struct {
	message    string
	optionPath []string
	cause      error
}

func (c configError) Error() string {
	return fmt.Sprintf("invalid configuration option: %s (%s)", strings.Join(c.optionPath, "."), c.message)
}

func (c configError) Unwrap() error {
	return c.cause
}
