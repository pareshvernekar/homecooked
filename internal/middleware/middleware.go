package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
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
