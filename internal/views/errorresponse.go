package views

import (
	"errors"
	"net/http"
	"time"
)

// ErrorResponse represents a standardized error response format
type ErrorResponse struct {
	Success    bool      `json:"success"`
	ErrorCode  string    `json:"error_code"`
	Message    string    `json:"message"`
	RequestID  string    `json:"request_id,omitempty"`
	Detail     any       `json:"details,omitempty"` // Can be map, array, or error message
	StatusCode int       `json:"-"`                 // Not sent in response (used internally)
	Timestamp  time.Time `json:"timestamp"`
}

// Error implements the error interface for ErrorResponse
func (e ErrorResponse) Error() string {
	return e.Message
}

// ErrorCode represents predefined error codes for the application
type ErrorCode string

// Predefined error codes - use these constants instead of raw strings
const (
	SUCCESS                   ErrorCode = "SUCCESS"
	INVALID_REQUEST           ErrorCode = "INVALID_REQUEST"
	VALIDATION_ERROR          ErrorCode = "VALIDATION_ERROR"
	BAD_USER_INPUT            ErrorCode = "BAD_USER_INPUT"
	UNAUTHORIZED              ErrorCode = "UNAUTHORIZED"
	FORBIDDEN                 ErrorCode = "FORBIDDEN"
	NOT_FOUND                 ErrorCode = "NOT_FOUND"
	ALREADY_EXISTS            ErrorCode = "ALREADY_EXISTS"
	CONFLICT                  ErrorCode = "CONFLICT"
	BAD_GATEWAY               ErrorCode = "BAD_GATEWAY"
	SERVICE_UNAVAILABLE       ErrorCode = "SERVICE_UNAVAILABLE"
	INVALID_TENANT_ID         ErrorCode = "INVALID_TENANT_ID"
	INVALID_UUID              ErrorCode = "INVALID_UUID"
	ENTITY_NOT_FOUND          ErrorCode = "ENTITY_NOT_FOUND"
	DATABASE_ERROR            ErrorCode = "DATABASE_ERROR"
	CONNECTION_TIMEOUT        ErrorCode = "CONNECTION_TIMEOUT"
	INTERNAL_SERVER_ERROR     ErrorCode = "INTERNAL_SERVER_ERROR"
	RATE_LIMIT_EXCEEDED       ErrorCode = "RATE_LIMIT_EXCEEDED"
	INVALID_EMAIL_FORMAT      ErrorCode = "INVALID_EMAIL_FORMAT"
	INVALID_PHONE_FORMAT      ErrorCode = "INVALID_PHONE_FORMAT"
	INVALID_DATE_RANGE        ErrorCode = "INVALID_DATE_RANGE"
	INSUFFICIENT_PERMISSIONS  ErrorCode = "INSUFFICIENT_PERMISSIONS"
	TENANT_NOT_FOUND          ErrorCode = "TENANT_NOT_FOUND"
	ADMIN_ONLY_OPERATION      ErrorCode = "ADMIN_ONLY_OPERATION"
	COULD_NOT_CREATE_DATABASE ErrorCode = "COULD_NOT_CREATE_DATABASE"
)

func (e ErrorCode) String() string {
	return string(e)
}

// ValidationErrorDetail represents a single validation error detail
type ValidationErrorDetail struct {
	Field      string      `json:"field"`
	Code       string      `json:"code"`
	Message    string      `json:"message"`
	Value      interface{} `json:"value,omitempty"`      // Actual invalid value (if any)
	Constraint string      `json:"constraint,omitempty"` // Constraint that was violated
}

// ValidationError represents multiple validation errors in a request
type ValidationError struct {
	Success   bool                    `json:"success"`
	ErrorCode ErrorCode               `json:"error_code"`
	Message   string                  `json:"message"`
	Details   []ValidationErrorDetail `json:"details,omitempty"`
	RequestID string                  `json:"request_id,omitempty"`
	Timestamp time.Time               `json:"timestamp"`
}

// Error implements the error interface for ValidationError
func (e ValidationError) Error() string {
	return e.Message
}

// CreateErrorResponse creates a standardized error response
func CreateErrorResponse(code ErrorCode, message string, details interface{}, statusCode int) ErrorResponse {
	return ErrorResponse{
		Success:    false,
		ErrorCode:  code.String(),
		Message:    message,
		Detail:     details,
		StatusCode: statusCode,
		Timestamp:  time.Now().UTC(),
		RequestID:  "", // Set via middleware if available
	}
}

// CreateValidationError creates a validation error response with multiple field errors
func CreateValidationError(code ErrorCode, message string, details []ValidationErrorDetail) ValidationError {
	return ValidationError{
		Success:   false,
		ErrorCode: code,
		Message:   message,
		Details:   details,
		Timestamp: time.Now().UTC(),
		RequestID: "", // Set via middleware if available
	}
}

// Error implements the error interface for SuccessResponse (unused but for completeness)
func (s SuccessResponse) Error() string {
	return s.Message
}

// Helper function to create validation errors from field map
func CreateValidationErrors(fields map[string]string) []ValidationErrorDetail {
	var details []ValidationErrorDetail

	for field, msg := range fields {
		details = append(details, ValidationErrorDetail{
			Field:   field,
			Code:    "REQUIRED", // Default code - can be customized
			Message: msg,
		})
	}

	return details
}

// Helper function to get status code from error code
func GetStatusFromErrorCode(code ErrorCode) int {
	switch code {
	case SUCCESS:
		return http.StatusOK
	case INVALID_REQUEST, BAD_USER_INPUT, VALIDATION_ERROR:
		return http.StatusBadRequest
	case UNAUTHORIZED:
		return http.StatusUnauthorized
	case FORBIDDEN, INSUFFICIENT_PERMISSIONS:
		return http.StatusForbidden
	case NOT_FOUND, ENTITY_NOT_FOUND:
		return http.StatusNotFound
	case ALREADY_EXISTS, CONFLICT:
		return http.StatusConflict
	case DATABASE_ERROR, CONNECTION_TIMEOUT:
		return http.StatusBadGateway
	case RATE_LIMIT_EXCEEDED:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}

// Helper function to log error with code
func LogError(err error, code ErrorCode) {
	if errors.Is(err, http.ErrServerClosed) || err == nil {
		return
	}

	// Standard error logging - can use logger package instead
	println("[ERROR] " + err.Error() + " | Code: " + code.String())
}
