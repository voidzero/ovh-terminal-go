// internal/api/errors.go
package api

import (
	"fmt"
	"strings"
)

// ErrorType represents different categories of API errors
type ErrorType int

const (
	ErrorTypeAuth ErrorType = iota
	ErrorTypeAPI
	ErrorTypeNetwork
	ErrorTypeValidation
)

// errorMessages maps error types to their user-friendly messages
var errorMessages = map[ErrorType]string{
	ErrorTypeAuth:       "Authentication failed. Please check your API credentials in the configuration file.",
	ErrorTypeAPI:        "The OVH API reported an error.",
	ErrorTypeNetwork:    "Could not connect to OVH API. Please check your internet connection.",
	ErrorTypeValidation: "Invalid request. This might be a bug in the application.",
}

// UserMessage returns a user-friendly message for each error type
func (et ErrorType) UserMessage() string {
	if msg, ok := errorMessages[et]; ok {
		return msg
	}
	return "An unknown error occurred."
}

// APIError wraps API-related errors with additional context
type APIError struct {
	Type    ErrorType
	Message string
	Details interface{}
	Err     error
}

func (e *APIError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// ErrorDetail represents structured error details
type ErrorDetail struct {
	Status int
	Method string
	Path   string
}

// detailsToString formats error details into a readable string
func detailsToString(details interface{}) string {
	if details == nil {
		return ""
	}

	if d, ok := details.(map[string]interface{}); ok {
		var parts []string
		if status, exists := d["status"]; exists {
			parts = append(parts, fmt.Sprintf("Status: %v", status))
		}
		if method, exists := d["method"]; exists {
			parts = append(parts, fmt.Sprintf("Method: %v", method))
		}
		if path, exists := d["path"]; exists {
			parts = append(parts, fmt.Sprintf("Path: %v", path))
		}
		if len(parts) > 0 {
			return fmt.Sprintf(" (%s)", strings.Join(parts, ", "))
		}
	}
	return ""
}

// UserError returns a user-friendly error message
func (e *APIError) UserError() string {
	baseMsg := e.Type.UserMessage()

	switch e.Type {
	case ErrorTypeAuth:
		if strings.Contains(strings.ToLower(e.Error()), "invalid") {
			return fmt.Sprintf("%s The credentials appear to be invalid.", baseMsg)
		}
		if strings.Contains(strings.ToLower(e.Error()), "expired") {
			return fmt.Sprintf("%s Your authentication token may have expired.", baseMsg)
		}
		return baseMsg

	case ErrorTypeAPI:
		details := detailsToString(e.Details)
		return baseMsg + details

	default:
		return baseMsg
	}
}

func (e *APIError) Unwrap() error {
	return e.Err
}

// Error constructors
func NewAuthError(message string, err error) *APIError {
	return &APIError{
		Type:    ErrorTypeAuth,
		Message: message,
		Err:     err,
	}
}

func NewAPIError(message string, err error, details interface{}) *APIError {
	return &APIError{
		Type:    ErrorTypeAPI,
		Message: message,
		Details: details,
		Err:     err,
	}
}

func NewNetworkError(message string, err error) *APIError {
	return &APIError{
		Type:    ErrorTypeNetwork,
		Message: message,
		Err:     err,
	}
}

func NewValidationError(message string, err error) *APIError {
	return &APIError{
		Type:    ErrorTypeValidation,
		Message: message,
		Err:     err,
	}
}

