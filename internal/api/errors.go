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

// UserMessage returns a user-friendly message for each error type
func (et ErrorType) UserMessage() string {
	switch et {
	case ErrorTypeAuth:
		return "Authentication failed. Please check your API credentials in the configuration file."
	case ErrorTypeAPI:
		return "The OVH API reported an error."
	case ErrorTypeNetwork:
		return "Could not connect to OVH API. Please check your internet connection."
	case ErrorTypeValidation:
		return "Invalid request. This might be a bug in the application."
	default:
		return "An unknown error occurred."
	}
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

// UserError returns a user-friendly error message
func (e *APIError) UserError() string {
	baseMsg := e.Type.UserMessage()

	// Add specific details based on error type
	switch e.Type {
	case ErrorTypeAuth:
		if strings.Contains(e.Error(), "invalid") {
			return fmt.Sprintf("%s\nThe credentials appear to be invalid.", baseMsg)
		}
		if strings.Contains(e.Error(), "expired") {
			return fmt.Sprintf("%s\nYour authentication token may have expired.", baseMsg)
		}
		return baseMsg

	case ErrorTypeAPI:
		if details, ok := e.Details.(map[string]interface{}); ok {
			if status, exists := details["status"]; exists {
				return fmt.Sprintf("%s (Status: %v)", baseMsg, status)
			}
		}
		return baseMsg

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

