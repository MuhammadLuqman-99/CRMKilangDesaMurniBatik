// Package common contains shared HTTP utilities.
package common

import (
	"encoding/json"
	"net/http"
	"time"
)

// Standard error codes.
const (
	ErrCodeBadRequest     = "BAD_REQUEST"
	ErrCodeUnauthorized   = "UNAUTHORIZED"
	ErrCodeForbidden      = "FORBIDDEN"
	ErrCodeNotFound       = "NOT_FOUND"
	ErrCodeConflict       = "CONFLICT"
	ErrCodeValidation     = "VALIDATION_ERROR"
	ErrCodeInternal       = "INTERNAL_ERROR"
	ErrCodeTooManyRequest = "TOO_MANY_REQUESTS"
)

// Response represents a standard API response.
type Response struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Meta      *MetaInfo   `json:"meta,omitempty"`
	Timestamp string      `json:"timestamp"`
}

// ErrorInfo represents error details in an API response.
type ErrorInfo struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// MetaInfo represents metadata in an API response.
type MetaInfo struct {
	Page       int   `json:"page,omitempty"`
	PageSize   int   `json:"page_size,omitempty"`
	TotalItems int64 `json:"total_items,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
}

// NewSuccessResponse creates a new success response.
func NewSuccessResponse(data interface{}) Response {
	return Response{
		Success:   true,
		Data:      data,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// NewSuccessResponseWithMeta creates a new success response with metadata.
func NewSuccessResponseWithMeta(data interface{}, meta *MetaInfo) Response {
	return Response{
		Success:   true,
		Data:      data,
		Meta:      meta,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// NewErrorResponse creates a new error response.
func NewErrorResponse(code, message string, details map[string]interface{}) Response {
	return Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// WriteJSON writes a JSON response.
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Log error but can't write again
		_ = err
	}
}

// WriteSuccess writes a success JSON response.
func WriteSuccess(w http.ResponseWriter, status int, data interface{}) {
	WriteJSON(w, status, NewSuccessResponse(data))
}

// WriteSuccessWithMeta writes a success JSON response with metadata.
func WriteSuccessWithMeta(w http.ResponseWriter, status int, data interface{}, meta *MetaInfo) {
	WriteJSON(w, status, NewSuccessResponseWithMeta(data, meta))
}

// WriteError writes an error JSON response.
func WriteError(w http.ResponseWriter, status int, code, message string, details map[string]interface{}) {
	WriteJSON(w, status, NewErrorResponse(code, message, details))
}

// CalculateTotalPages calculates total pages for pagination.
func CalculateTotalPages(total int64, pageSize int) int {
	if pageSize <= 0 {
		return 0
	}
	pages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		pages++
	}
	return pages
}
