package client

import (
	"fmt"
)

// Error codes from TrueNAS API
const (
	ErrCodeParseError     = -32700
	ErrCodeInvalidRequest = -32600
	ErrCodeMethodNotFound = -32601
	ErrCodeInvalidParams  = -32602
	ErrCodeInternalError  = -32603

	// TrueNAS specific error codes
	ErrCodeNotAuthenticated = 1
	ErrCodeNotAuthorized    = 2
	ErrCodeNotFound         = 3
	ErrCodeValidation       = 4
)

// APIError represents an error from the TrueNAS API
type APIError struct {
	Code    int
	Message string
	Details string
}

func (e *APIError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("TrueNAS API error %d: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("TrueNAS API error %d: %s", e.Code, e.Message)
}

// IsNotFound returns true if the error indicates a resource was not found
func (e *APIError) IsNotFound() bool {
	return e.Code == ErrCodeNotFound
}

// IsAuthError returns true if the error is an authentication error
func (e *APIError) IsAuthError() bool {
	return e.Code == ErrCodeNotAuthenticated || e.Code == ErrCodeNotAuthorized
}

// IsValidationError returns true if the error is a validation error
func (e *APIError) IsValidationError() bool {
	return e.Code == ErrCodeValidation
}

// ConnectionError represents a connection-related error
type ConnectionError struct {
	Message string
	Err     error
}

func (e *ConnectionError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("connection error: %s: %v", e.Message, e.Err)
	}
	return fmt.Sprintf("connection error: %s", e.Message)
}

func (e *ConnectionError) Unwrap() error {
	return e.Err
}

// NewAPIError creates a new APIError from a JSONRPCError
func NewAPIError(rpcErr *JSONRPCError) *APIError {
	details := ""
	if rpcErr.Data != nil {
		details = string(rpcErr.Data)
	}
	return &APIError{
		Code:    rpcErr.Code,
		Message: rpcErr.Message,
		Details: details,
	}
}

// NewConnectionError creates a new ConnectionError
func NewConnectionError(message string, err error) *ConnectionError {
	return &ConnectionError{
		Message: message,
		Err:     err,
	}
}

// IsNotFoundError checks if an error is a not found error
func IsNotFoundError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.IsNotFound()
	}
	return false
}

// IsAuthError checks if an error is an authentication error
func IsAuthError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.IsAuthError()
	}
	return false
}
