// Package response provides HTTP response utilities for the CRM application.
package response

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/kilang-desa-murni/crm/pkg/errors"
)

// Response represents a standard API response.
type Response struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorBody  `json:"error,omitempty"`
	Meta      *Meta       `json:"meta,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// ErrorBody represents the error details in a response.
type ErrorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details string            `json:"details,omitempty"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// Meta holds metadata for paginated responses.
type Meta struct {
	Page       int   `json:"page,omitempty"`
	PerPage    int   `json:"per_page,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
}

// PaginatedData represents paginated data with items.
type PaginatedData struct {
	Items interface{} `json:"items"`
}

// JSON writes a JSON response.
func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := Response{
		Success:   statusCode >= 200 && statusCode < 300,
		Data:      data,
		Timestamp: time.Now().UTC(),
	}

	json.NewEncoder(w).Encode(response)
}

// OK writes a 200 OK response.
func OK(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, data)
}

// Created writes a 201 Created response.
func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, data)
}

// NoContent writes a 204 No Content response.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Accepted writes a 202 Accepted response.
func Accepted(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusAccepted, data)
}

// Paginated writes a paginated response.
func Paginated(w http.ResponseWriter, items interface{}, page, perPage int, total int64) {
	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := Response{
		Success: true,
		Data: PaginatedData{
			Items: items,
		},
		Meta: &Meta{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		},
		Timestamp: time.Now().UTC(),
	}

	json.NewEncoder(w).Encode(response)
}

// Error writes an error response.
func Error(w http.ResponseWriter, err error) {
	var statusCode int
	var errorBody ErrorBody

	if appErr, ok := errors.AsAppError(err); ok {
		statusCode = appErr.HTTPStatus()
		errorBody = ErrorBody{
			Code:    string(appErr.Code),
			Message: appErr.Message,
			Details: appErr.Details,
			Fields:  appErr.Fields,
		}
	} else {
		statusCode = http.StatusInternalServerError
		errorBody = ErrorBody{
			Code:    string(errors.ErrCodeInternal),
			Message: "An internal error occurred",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := Response{
		Success:   false,
		Error:     &errorBody,
		Timestamp: time.Now().UTC(),
	}

	json.NewEncoder(w).Encode(response)
}

// BadRequest writes a 400 Bad Request error response.
func BadRequest(w http.ResponseWriter, message string) {
	Error(w, errors.ErrBadRequest(message))
}

// Unauthorized writes a 401 Unauthorized error response.
func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, errors.ErrUnauthorized(message))
}

// Forbidden writes a 403 Forbidden error response.
func Forbidden(w http.ResponseWriter, message string) {
	Error(w, errors.ErrForbidden(message))
}

// NotFound writes a 404 Not Found error response.
func NotFound(w http.ResponseWriter, resource string) {
	Error(w, errors.ErrNotFound(resource))
}

// Conflict writes a 409 Conflict error response.
func Conflict(w http.ResponseWriter, message string) {
	Error(w, errors.ErrConflict(message))
}

// InternalError writes a 500 Internal Server Error response.
func InternalError(w http.ResponseWriter, message string) {
	Error(w, errors.ErrInternal(message))
}

// ValidationError writes a validation error response.
func ValidationError(w http.ResponseWriter, fields map[string]string) {
	appErr := errors.ErrValidation("Validation failed").WithFields(fields)
	Error(w, appErr)
}

// HealthResponse represents a health check response.
type HealthResponse struct {
	Status    string                 `json:"status"`
	Version   string                 `json:"version,omitempty"`
	Uptime    string                 `json:"uptime,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Checks    map[string]HealthCheck `json:"checks,omitempty"`
}

// HealthCheck represents an individual health check.
type HealthCheck struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// Health writes a health check response.
func Health(w http.ResponseWriter, status string, version string, uptime time.Duration, checks map[string]HealthCheck) {
	statusCode := http.StatusOK
	if status != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := HealthResponse{
		Status:    status,
		Version:   version,
		Uptime:    uptime.String(),
		Timestamp: time.Now().UTC(),
		Checks:    checks,
	}

	json.NewEncoder(w).Encode(response)
}

// Stream writes a streaming response with Server-Sent Events.
func Stream(w http.ResponseWriter, eventChan <-chan interface{}, done <-chan struct{}) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	for {
		select {
		case <-done:
			return
		case event, ok := <-eventChan:
			if !ok {
				return
			}
			data, err := json.Marshal(event)
			if err != nil {
				continue
			}
			w.Write([]byte("data: "))
			w.Write(data)
			w.Write([]byte("\n\n"))
			flusher.Flush()
		}
	}
}

// Download writes a file download response.
func Download(w http.ResponseWriter, filename string, contentType string, data []byte) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.Header().Set("Content-Length", string(rune(len(data))))
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
