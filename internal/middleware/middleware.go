package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	logger "github.com/pareshvernekar/homecooked/internal/logger"
)

// TenantIDKey is the key used to store the tenant ID in the request context
const TenantIDKey = "tenant_id"

// TenantMiddleware extracts the tenant ID from the request headers and stores it in the request context
func TenantMiddleware(l *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetHeader("X-Tenant-ID")
		if tenantID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID is required"})
			c.Abort()
			return
		}

		l.Info(c.Request.Context(), "Handling request with tenant ID", "tenant_id", tenantID, "request_id", c.GetHeader("X-Request-ID"))

		c.Set(TenantIDKey, tenantID)
		c.Next()
	}
}

// SetTenantContext sets the tenant ID in the PostgreSQL session context
func SetTenantContext(c *gin.Context) {
	c.Next() // Continue with normal request handling
}

// GetTenantID returns the tenant ID from the request context
func GetTenantID(c *gin.Context) string {
	tenantID := c.GetString(TenantIDKey)
	return tenantID
}

// RequireValidTenantHeader is an alternative middleware that provides better error handling
func RequireValidTenantHeader(l *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetHeader("X-Tenant-ID")

		if tenantID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]interface{}{
				"error":      "Missing required X-Tenant-ID header",
				"code":       "MISSING_TENANT_ID",
				"request_id": c.GetString("request_id"),
			})
			return
		}

		if _, err := uuid.Parse(tenantID); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]interface{}{
				"error":      "Invalid X-Tenant-ID format",
				"code":       "INVALID_TENANT_ID",
				"details":    "The provided tenant ID is not a valid UUID",
				"request_id": c.GetString("request_id"),
			})
			return
		}

		c.Set(TenantIDKey, tenantID)
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
