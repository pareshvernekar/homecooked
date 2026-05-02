package main

import (
	"github.com/gin-gonic/gin"
	"github.com/pareshvernekar/homecooked/internal/config"
	"github.com/pareshvernekar/homecooked/internal/middleware"
	"github.com/pareshvernekar/homecooked/internal/logger"
	"log/slog"
	"net/http"
)

func main() {
	// Initialize the configuration
	config.Init()
	
	// Create the Gin router
	router := gin.Default()

	// Use the TenantMiddleware
	router.Use(middleware.TenantMiddleware())

	// Define a route that uses the tenant ID from the context
	router.GET("/", func(c *gin.Context) {
		// Retrieve the tenant ID from the context
		tenantID, exists := c.Get(middleware.TenantIDKey)
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Tenant ID not found"})
			return
		}
		// Log the request with the tenant ID
		logger.Logger.Info("Handling request with tenant ID", slog.Any("tenant_id", tenantID), slog.String("request_id", c.GetHeader("X-Request-ID")))
		
		// Respond with the tenant ID
		c.JSON(http.StatusOK, gin.H{"tenant_id": tenantID})
	})

	// Start the HTTP server
	router.Run(":8082")
}