package errors

import "fmt"

// Common error types
const (
	ErrInvalidPackageJSON  = "invalid package.json"
	ErrPackageJSONNotFound = "package.json not found"
	ErrInvalidCommand      = "invalid command"
)

// ZapError represents a custom error type for Zap
type ZapError struct {
	Type    string
	Message string
	Err     error
}

// Error implements the error interface
func (e *ZapError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// New creates a new ZapError
func New(errType string, message string, err error) *ZapError {
	return &ZapError{
		Type:    errType,
		Message: message,
		Err:     err,
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	if zapErr, ok := err.(*ZapError); ok {
		return &ZapError{
			Type:    zapErr.Type,
			Message: fmt.Sprintf("%s: %s", message, zapErr.Message),
			Err:     zapErr.Err,
		}
	}
	return &ZapError{
		Type:    "unknown",
		Message: message,
		Err:     err,
	}
}
