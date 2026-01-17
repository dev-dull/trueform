package client

import (
	"fmt"
	"strings"
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
	if e.Code == ErrCodeNotFound {
		return true
	}
	// TrueNAS Scale 25 returns InvalidParams with InstanceNotFound in details
	if e.Code == ErrCodeInvalidParams {
		if strings.Contains(e.Details, "InstanceNotFound") || strings.Contains(e.Details, "does not exist") {
			return true
		}
	}
	return false
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
	Host string
	Err  error
}

func (e *ConnectionError) Error() string {
	return fmt.Sprintf(`failed to connect to TrueNAS at %q: %v

Please verify:
  1. The host is reachable (try: curl -k https://%s/api/current)
  2. TrueNAS Scale 25.04+ is running and the API is enabled
  3. Your provider configuration is correct

Example configuration:

  provider "trueform" {
    host       = "192.168.1.100"    # TrueNAS IP or hostname
    api_key    = "1-xxxx..."        # API key from TrueNAS UI
    verify_ssl = false              # Set true if using valid SSL cert
  }

Or use environment variables:
  export TRUENAS_HOST="192.168.1.100"
  export TRUENAS_API_KEY="1-xxxx..."
  export TRUENAS_VERIFY_SSL="false"
`, e.Host, e.Err, e.Host)
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
func NewConnectionError(host string, err error) *ConnectionError {
	return &ConnectionError{
		Host: host,
		Err:  err,
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
