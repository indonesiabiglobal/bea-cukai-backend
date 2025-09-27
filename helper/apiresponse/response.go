package apiresponse

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// APIError holds machine-friendly error info
// Code: a short string you define per failure path (e.g., "KPI_FETCH_FAILED")
// Details: error detail; hidden in production
type APIError struct {
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// Response is a generic API wrapper for consistent outputs
// Success: true/false to simplify FE logic
// Message: short, human-friendly summary (e.g., "ok")
// Data: payload (any shape)
// Meta: optional metadata (e.g., pagination, filters, date range)
// Error: populated only when Success=false
// TraceID: request correlation id; filled if request id middleware/header is present
type Response[T any] struct {
	Message string    `json:"message,omitempty"`
	Data    *T        `json:"data,omitempty"`
	Meta    any       `json:"meta,omitempty"`
	Error   *APIError `json:"error,omitempty"`
	TraceID string    `json:"trace_id,omitempty"`
}

// PageMeta is a common pagination block for list endpoints
type PageMeta struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}

// getTraceID tries common places: writer header, request header, gin context key
func getTraceID(c *gin.Context) string {
	if v := c.Writer.Header().Get("X-Request-ID"); v != "" {
		return v
	}
	if v := c.GetHeader("X-Request-ID"); v != "" {
		// bubble up to response header as well
		c.Writer.Header().Set("X-Request-ID", v)
		return v
	}
	if v, ok := c.Get("requestid"); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func isProd() bool {
	return os.Getenv("APP_ENV") == "production" || os.Getenv("GIN_MODE") == "release"
}

// OK writes a 200 response with a wrapped payload
func OK[T any](c *gin.Context, data T, message string, meta any) {
	resp := Response[T]{
		Message: message,
		Data:    &data,
		Meta:    meta,
		TraceID: getTraceID(c),
	}
	c.JSON(http.StatusOK, resp)
}

// Created writes a 201 response with payload
func Created[T any](c *gin.Context, data T, message string, meta any) {
	resp := Response[T]{
		Message: message,
		Data:    &data,
		Meta:    meta,
		TraceID: getTraceID(c),
	}
	c.JSON(http.StatusCreated, resp)
}

const (
	ErrCodeBadRequest = "BAD_REQUEST"
	ErrCodeInternal   = "INTERNAL_ERROR"
)

// Error writes a JSON error with consistent body and status code
// If not production, it includes err.Error() in Details
func Error(c *gin.Context, status int, code, message string, err error, meta any) {
	apiErr := &APIError{Code: code}
	if err != nil && !isProd() {
		apiErr.Details = err.Error()
	}
	resp := Response[any]{
		Message: message,
		Error:   apiErr,
		Meta:    meta,
		TraceID: getTraceID(c),
	}
	c.JSON(status, resp)
}

// Versi dengan code & meta kustom (opsional)
func BadRequest(c *gin.Context, code, message string, err error, meta any) {
	Error(c, http.StatusBadRequest, code, message, err, meta)
}

func InternalServerError(c *gin.Context, code, message string, err error, meta any) {
	Error(c, http.StatusInternalServerError, code, message, err, meta)
}
