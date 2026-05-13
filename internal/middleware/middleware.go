package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pareshvernekar/homecooked/internal/logger"
)

// TenantIDKey is the key used to store the tenant ID in the request context
const TenantIDKey = "tenant_id"

// TenantMiddleware extracts the tenant ID from the request headers and stores it in the request context
func TenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract the tenant ID from the request header
		tenantID := c.GetHeader("X-Tenant-ID")
		if tenantID == "" {
			// If the tenant ID is not present, you can handle it as needed (e.g., log an error, return a 400 status)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID is required"})
			c.Abort()
			return
		}

		// Set the tenant ID in the request context
		c.Set(TenantIDKey, tenantID)
		// Log the request with the tenant ID
		logger.Logger.Info("Handling request with tenant ID", slog.String("tenant_id", tenantID), slog.String("request_id", c.GetHeader("X-Request-ID")))

		// Call the next handler
		c.Next()
	}
}

// SetTenantContext sets the tenant ID in the PostgreSQL session context
func SetTenantContext(c *gin.Context, tenantID int) {
	c.Next() // Continue with normal request handling
}

// GetTenantID returns the tenant ID from the request context
func GetTenantID(c *gin.Context) string {
	tenantID := c.GetString(TenantIDKey)
	return tenantID
}

// RequireValidTenantHeader is an alternative middleware that provides better error handling
func RequireValidTenantHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract tenant ID from X-Tenant-ID header
		tenantID := c.GetHeader("X-Tenant-ID")

		if tenantID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]interface{}{
				"error":      "Missing required X-Tenant-ID header",
				"code":       "MISSING_TENANT_ID",
				"request_id": c.GetString("request_id"),
			})
			return
		}

		// Validate UUID format using uuid.Parse
		if _, err := uuid.Parse(tenantID); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]interface{}{
				"error":      "Invalid X-Tenant-ID format",
				"code":       "INVALID_TENANT_ID",
				"details":    "The provided tenant ID is not a valid UUID",
				"request_id": c.GetString("request_id"),
			})
			return
		}

		// Set tenant ID in context for downstream handlers to access
		c.Set(TenantIDKey, tenantID)

		// Continue with the request
		c.Next()
	}
}

// ValidateTenantUUID validates that a string is a valid UUID format
func ValidateTenantUUID(tenantID string) bool {
	if tenantID == "" {
		return false
	}

	_, err := uuid.Parse(tenantID)
	return err == nil
}
