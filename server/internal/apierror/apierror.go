// Package apierror defines the structured error type and sentinel errors used
// to signal well-known failure conditions (already exists, not found, parent
// resource missing) from the service layer up to the HTTP layer.
package apierror

import "errors"

// APIError represents a structured error for API responses.
// Includes a code and message for consistent error handling.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error implements the error interface for APIError.
func (e *APIError) Error() string {
	return e.Message
}

// NewAPIError creates a new APIError with the given code and message.
func NewAPIError(code, message string) *APIError {
	return &APIError{Code: code, Message: message}
}

// NewValidationError creates a plain error for a request-validation failure.
func NewValidationError(message string) error {
	return errors.New(message)
}

// Sentinel errors services return for well-known failure conditions.
// Handlers match against these with errors.Is/errors.As to pick an HTTP
// status code.
var (
	ErrResourceAlreadyExist = NewAPIError("RESOURCE_ALREADY_EXISTS", "resource already exists")
	ErrResourceNotFound     = NewAPIError("RESOURCE_DOES_NOT_EXIST", "resource doesnt exist")
	ErrParentNotExist       = NewAPIError("PARENT_RESOURCE_DOES_NOT_EXIST", "parent resource  not exist")
	// ErrNotNullViolation and ErrTransactionConflict are internal-only:
	// handlers must never serialize their Code/Message directly to a
	// client, only use them for errors.Is branching and logging.
	ErrNotNullViolation    = NewAPIError("NOT_NULL_VIOLATION", "a required value was missing")
	ErrTransactionConflict = NewAPIError("TRANSACTION_CONFLICT", "transaction conflict, please retry")
)
