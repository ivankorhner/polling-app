package server

import (
	"encoding/json"
	"net/http"
)

// MaxRequestBodySize is the maximum allowed size for request bodies (1MB)
const MaxRequestBodySize = 1 << 20

// ErrorResponse represents a unified JSON error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// Error codes for common error scenarios
const (
	ErrCodeValidation   = "VALIDATION_ERROR"
	ErrCodeNotFound     = "NOT_FOUND"
	ErrCodeConflict     = "CONFLICT"
	ErrCodeInternal     = "INTERNAL_ERROR"
	ErrCodeBadRequest   = "BAD_REQUEST"
	ErrCodeUnauthorized = "UNAUTHORIZED"
)

// writeError writes a JSON error response
func writeError(w http.ResponseWriter, message, code string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{
		Error: message,
		Code:  code,
	})
}

// writeValidationError writes a validation error response
func writeValidationError(w http.ResponseWriter, message string) {
	writeError(w, message, ErrCodeValidation, http.StatusBadRequest)
}

// writeNotFoundError writes a not found error response
func writeNotFoundError(w http.ResponseWriter, message string) {
	writeError(w, message, ErrCodeNotFound, http.StatusNotFound)
}

// writeConflictError writes a conflict error response
func writeConflictError(w http.ResponseWriter, message string) {
	writeError(w, message, ErrCodeConflict, http.StatusConflict)
}

// writeInternalError writes an internal server error response
func writeInternalError(w http.ResponseWriter, message string) {
	writeError(w, message, ErrCodeInternal, http.StatusInternalServerError)
}
