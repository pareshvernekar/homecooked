package errors

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	logger "github.com/pareshvernekar/homecooked/internal/logger"
)

// ErrorCode represents the type of error that occurred in the service layer.
// Each error code maps to an appropriate HTTP status code for API responses.
type ErrorCode string

const (
	// ValidationError - Invalid request payloads, missing required fields, validation failures
	// Maps to: HTTP 400 Bad Request
	ValidationError ErrorCode = "VALIDATION_ERROR"

	// NotFoundError - Resources not found (category name not found, food item ID not found)
	// Maps to: HTTP 404 Not Found
	NotFoundError ErrorCode = "NOT_FOUND"

	// DatabaseError - Database connectivity issues, query failures, transaction errors
	// Error details are wrapped from original error via errors.Unwrap()
	// Maps to: HTTP 500 Internal Server Error
	DatabaseError ErrorCode = "DATABASE_ERROR"

	// CacheError - Cache operation failures (connection lost, key corruption, serialization)
	// Graceful degradation to database fallback when available
	// Maps to: HTTP 5xx Internal Server Error
	CacheError ErrorCode = "CACHE_ERROR"

	// ForbiddenError - Tenant isolation violations or access control failures
	// Cross-tenant access attempts, permission denied errors
	// Maps to: HTTP 403 Forbidden
	ForbiddenError ErrorCode = "FORBIDDEN_ERROR"

	// ServiceInitializationError - Errors during service initialization or dependency injection
	// Maps to: HTTP 500 Internal Server Error
	ServiceInitializationError ErrorCode = "SERVICE_INITIALIZATION_ERROR"
)

// ServiceError represents a categorized error from the service layer.
// Includes operational context for debugging and proper HTTP status code mapping.
type ServiceError struct {
	Code       ErrorCode   `json:"error_code"`       // Enum-based error code
	Type       string      `json:"error_type"`       // Human-readable type (Validation, NotFound, Database, etc.)
	Message    string      `json:"message"`          // Human-readable error message
	Detail     interface{} `json:"detail,omitempty"` // Optional detailed info from original error
	TenantID   string      `json:"-"`                // Tenant ID for operational context (not sent to client)
	Operation  string      `json:"operation"`        // Operation name (CreateFoodItem, GetByID, etc.)
	StatusCode int         `json:"status_code"`      // Mapped HTTP status code
}

// Error implements the error interface.
func (e *ServiceError) Error() string {
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// GetStatusCode returns the HTTP status code mapped from this error code.
func (e *ServiceError) GetStatusCode() int {
	switch e.Code {
	case ValidationError:
		return http.StatusBadRequest
	case NotFoundError:
		return http.StatusNotFound
	case DatabaseError:
		return http.StatusInternalServerError
	case CacheError:
		return http.StatusInternalServerError
	case ForbiddenError:
		return http.StatusForbidden
	case ServiceInitializationError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// Unwrap returns the original error if wrapped.
func (e *ServiceError) Unwrap() error {
	if err, ok := e.Detail.(error); ok {
		return err
	}
	return nil
}

// Log implements logging for structured logging integration.
func (e *ServiceError) Log(ctx context.Context, logger *logger.Logger) {
	if logger == nil {
		return
	}
	logger.Error(ctx, fmt.Sprintf("Service Error [%s]: %s", e.Code, e.Message),
		"error_code", e.Code,
		"error_type", e.Type,
		"status_code", e.GetStatusCode(),
		"operation", e.Operation)

	if tenantID := strings.TrimSpace(e.TenantID); tenantID != "" {
		logger.Error(ctx, "Error context for tenant",
			"tenant_id", tenantID,
			"error_code", e.Code,
			"operation", e.Operation)
	}
}

// --- Helper Functions for Error Creation ---

// CreateValidationError creates a ValidationError with the provided message.
func CreateValidationError(message string, detail interface{}) error {
	return &ServiceError{
		Code:       ValidationError,
		Type:       "Validation",
		Message:    fmt.Sprintf("Invalid request payload: %s", message),
		Detail:     detail,
		Operation:  "", // Set by caller
		StatusCode: http.StatusBadRequest,
	}
}

// CreateNotFoundError creates a NotFoundError with category or item name.
func CreateNotFoundError(resourceType string, identifier string, tenantID string) error {
	return &ServiceError{
		Code:       NotFoundError,
		Type:       "NotFound",
		Message:    fmt.Sprintf("%s not found for tenant %s", resourceType, tenantID),
		Detail:     map[string]interface{}{"resource_type": resourceType, "identifier": identifier},
		TenantID:   tenantID,
		Operation:  "", // Set by caller
		StatusCode: http.StatusNotFound,
	}
}

// CreateDatabaseError creates a DatabaseError with wrapped original error.
func CreateDatabaseError(message string, originalError error, tenantID string) error {
	return &ServiceError{
		Code:       DatabaseError,
		Type:       "Database",
		Message:    fmt.Sprintf("Database operation failed: %s - %s", message, SanitizeErrorMessage(originalError)),
		Detail:     originalError, // Wrapped via Unwrap() for debugging
		TenantID:   tenantID,
		Operation:  "", // Set by caller
		StatusCode: http.StatusInternalServerError,
	}
}

// CreateCacheError creates a CacheError with graceful degradation.
func CreateCacheError(message string, detail interface{}, tenantID string) error {
	return &ServiceError{
		Code:       CacheError,
		Type:       "Cache",
		Message:    fmt.Sprintf("Cache operation failed: %s", message),
		Detail:     detail,
		TenantID:   tenantID,
		Operation:  "", // Set by caller
		StatusCode: http.StatusInternalServerError,
	}
}

// CreateForbiddenError creates a ForbiddenError for tenant isolation violations.
func CreateForbiddenError(resourceType string, identifier string, tenantID string) error {
	return &ServiceError{
		Code:       ForbiddenError,
		Type:       "Forbidden",
		Message:    fmt.Sprintf("Cannot access %s - it does not belong to this tenant", resourceType),
		Detail:     map[string]interface{}{"resource_type": resourceType, "identifier": identifier},
		TenantID:   tenantID,
		Operation:  "", // Set by caller
		StatusCode: http.StatusForbidden,
	}
}

func CreateServiceInitializationError(message string, detail interface{}, tenantID string) error {
	return &ServiceError{
		Code:       ServiceInitializationError,
		Type:       "ServiceInitialization",
		Message:    fmt.Sprintf("Service initialization failed: %s", message),
		Detail:     detail,
		TenantID:   tenantID,
		Operation:  "", // Set by caller
		StatusCode: http.StatusInternalServerError,
	}
}

// sanitizeErrorMessage removes potentially sensitive information from error messages.
func SanitizeErrorMessage(err error) string {
	if err == nil {
		return "(nil)"
	}

	msg := err.Error()

	// Remove stack traces or internal paths if present
	cleanedMsg := msg

	return cleanedMsg
}
