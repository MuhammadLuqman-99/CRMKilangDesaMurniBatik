// Package common contains shared HTTP utilities.
package common

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// RequestDecoder decodes and validates HTTP request bodies.
type RequestDecoder struct {
	validator *validator.Validate
	maxSize   int64
}

// NewRequestDecoder creates a new request decoder.
func NewRequestDecoder() *RequestDecoder {
	v := validator.New()

	// Register custom validators
	v.RegisterValidation("slug", validateSlug)
	v.RegisterValidation("permission", validatePermission)

	return &RequestDecoder{
		validator: v,
		maxSize:   1 << 20, // 1MB
	}
}

// Decode decodes and validates a request body.
func (d *RequestDecoder) Decode(r *http.Request, dest interface{}) error {
	// Limit body size
	r.Body = http.MaxBytesReader(nil, r.Body, d.maxSize)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dest); err != nil {
		return fmt.Errorf("invalid request body: %w", err)
	}

	if err := d.validator.Struct(dest); err != nil {
		return err
	}

	return nil
}

// Validate validates a struct.
func (d *RequestDecoder) Validate(s interface{}) error {
	return d.validator.Struct(s)
}

// Custom validators

func validateSlug(fl validator.FieldLevel) bool {
	slug := fl.Field().String()
	if len(slug) == 0 || len(slug) > 100 {
		return false
	}

	for _, c := range slug {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return false
		}
	}

	return slug[0] != '-' && slug[len(slug)-1] != '-'
}

func validatePermission(fl validator.FieldLevel) bool {
	perm := fl.Field().String()
	parts := strings.Split(perm, ":")
	return len(parts) == 2 && len(parts[0]) > 0 && len(parts[1]) > 0
}

// ValidationErrors converts validator errors to a map.
func ValidationErrors(err error) map[string]interface{} {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		errors := make(map[string]interface{})
		for _, e := range validationErrors {
			field := strings.ToLower(e.Field())
			errors[field] = fmt.Sprintf("failed on '%s' validation", e.Tag())
		}
		return errors
	}
	return map[string]interface{}{"error": err.Error()}
}

// QueryParams helps parse query parameters.
type QueryParams struct {
	r *http.Request
}

// NewQueryParams creates a new query params helper.
func NewQueryParams(r *http.Request) *QueryParams {
	return &QueryParams{r: r}
}

// String gets a string query parameter.
func (q *QueryParams) String(key, defaultValue string) string {
	value := q.r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// Int gets an integer query parameter.
func (q *QueryParams) Int(key string, defaultValue int) int {
	value := q.r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

// Int64 gets an int64 query parameter.
func (q *QueryParams) Int64(key string, defaultValue int64) int64 {
	value := q.r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return defaultValue
	}
	return intValue
}

// Bool gets a boolean query parameter.
func (q *QueryParams) Bool(key string, defaultValue bool) bool {
	value := q.r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return boolValue
}

// UUID gets a UUID query parameter.
func (q *QueryParams) UUID(key string) (uuid.UUID, error) {
	value := q.r.URL.Query().Get(key)
	if value == "" {
		return uuid.Nil, fmt.Errorf("missing parameter: %s", key)
	}

	return uuid.Parse(value)
}

// UUIDOptional gets an optional UUID query parameter.
func (q *QueryParams) UUIDOptional(key string) *uuid.UUID {
	value := q.r.URL.Query().Get(key)
	if value == "" {
		return nil
	}

	id, err := uuid.Parse(value)
	if err != nil {
		return nil
	}
	return &id
}

// StringSlice gets a comma-separated string slice query parameter.
func (q *QueryParams) StringSlice(key string) []string {
	value := q.r.URL.Query().Get(key)
	if value == "" {
		return nil
	}

	return strings.Split(value, ",")
}

// PaginationParams represents pagination parameters.
type PaginationParams struct {
	Page          int
	PageSize      int
	SortBy        string
	SortDirection string
}

// GetPagination extracts pagination parameters from request.
func (q *QueryParams) GetPagination() PaginationParams {
	page := q.Int("page", 1)
	if page < 1 {
		page = 1
	}

	pageSize := q.Int("page_size", 20)
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	sortBy := q.String("sort_by", "created_at")
	sortDirection := q.String("sort_direction", "desc")
	sortDirection = strings.ToUpper(sortDirection)
	if sortDirection != "ASC" && sortDirection != "DESC" {
		sortDirection = "DESC"
	}

	return PaginationParams{
		Page:          page,
		PageSize:      pageSize,
		SortBy:        sortBy,
		SortDirection: sortDirection,
	}
}

// PathParams helps extract path parameters.
type PathParams struct {
	params map[string]string
}

// NewPathParams creates a new path params helper.
func NewPathParams(params map[string]string) *PathParams {
	return &PathParams{params: params}
}

// String gets a string path parameter.
func (p *PathParams) String(key string) string {
	return p.params[key]
}

// UUID gets a UUID path parameter.
func (p *PathParams) UUID(key string) (uuid.UUID, error) {
	value := p.params[key]
	if value == "" {
		return uuid.Nil, fmt.Errorf("missing path parameter: %s", key)
	}
	return uuid.Parse(value)
}
