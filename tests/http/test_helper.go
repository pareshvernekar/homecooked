package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pareshvernekar/homecooked/internal/middleware"
)

// Response represents a standard API response structure
type Response struct {
	Success   bool        `json:"success"`
	ErrorCode string      `json:"error_code,omitempty"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// ErrorResponse represents an error response structure
type ErrorResponse struct {
	Success   bool        `json:"success"`
	ErrorCode string      `json:"error_code"`
	Message   string      `json:"message"`
	Detail    interface{} `json:"details,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// SuccessResponse represents a successful API response structure
type SuccessResponse struct {
	Success   bool            `json:"success"`
	Data      interface{}     `json:"data,omitempty"`
	Message   string          `json:"message,omitempty"`
	Paginated *PaginationInfo `json:"pagination,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
}

// PaginationInfo represents pagination information for paginated responses
type PaginationInfo struct {
	Total       int64 `json:"total"`
	Page        int   `json:"page"`
	Limit       int   `json:"limit"`
	HasNextPage bool  `json:"has_next_page"`
	HasPrevPage bool  `json:"has_prev_page"`
}

// Request represents a testable HTTP request structure
type Request struct {
	Method  string                 `json:"method"`
	Path    string                 `json:"path"`
	Body    map[string]interface{} `json:"body,omitempty"`
	Headers map[string]string      `json:"headers,omitempty"`
	Query   map[string]string      `json:"query,omitempty"`
}

// TestRequest creates a new HTTP test request with optional body and headers
func TestRequest(method string, path string, body map[string]interface{}, extraHeaders map[string]string) *http.Request {
	var bodyStr string
	if len(body) > 0 {
		var b strings.Builder
		err := json.NewEncoder(&b).Encode(body)
		if err != nil {
			panic(err)
		}
		bodyStr = b.String()
	}

	req := httptest.NewRequest(method, path, strings.NewReader(bodyStr))

	for k, v := range req.Header {
		req.Header.Add(k, v[0])
	}

	for key, value := range extraHeaders {
		req.Header.Set(key, value)
	}

	return req
}

// CreateTestContext creates a new test context from an HTTP request with tenant ID
func CreateTestContext(req *http.Request) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c := &gin.Context{
		Request: req,
	}

	// Set default tenant ID if not already set
	if req.Header.Get("X-Tenant-ID") == "" {
		req.Header.Set("X-Tenant-ID", "test-tenant")
	}
	c.Set(middleware.TenantIDKey, "test-tenant")
	return c, w
}

// Chains returns middleware chains from the handlers package
func Chains() []gin.HandlerFunc {
	return []gin.HandlerFunc{} // Placeholder - no middleware by default
}
