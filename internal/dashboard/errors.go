package dashboard

import (
	"encoding/json"
	"net/http"
)

// ErrorCode represents an error code
type ErrorCode string

const (
	ErrBadRequest       ErrorCode = "bad_request"
	ErrNotFound         ErrorCode = "not_found"
	ErrMethodNotAllowed ErrorCode = "method_not_allowed"
	ErrRequestTooLarge  ErrorCode = "request_too_large"
	ErrInternal         ErrorCode = "internal_error"
)

// APIError represents an API error response
type APIError struct {
	Code    ErrorCode `json:"error"`
	Message string    `json:"message"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	return string(e.Code) + ": " + e.Message
}

// WriteError writes an error response
func WriteError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(APIError{
		Code:    errorCodeForStatus(status),
		Message: message,
	})
}

// errorCodeForStatus maps HTTP status codes to error codes
func errorCodeForStatus(status int) ErrorCode {
	switch status {
	case http.StatusBadRequest:
		return ErrBadRequest
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusMethodNotAllowed:
		return ErrMethodNotAllowed
	case http.StatusRequestEntityTooLarge:
		return ErrRequestTooLarge
	case http.StatusInternalServerError:
		return ErrInternal
	default:
		return ErrInternal
	}
}

// WriteJSON writes a JSON response
func WriteJSON(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}
