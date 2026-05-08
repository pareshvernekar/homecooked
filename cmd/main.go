package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pareshvernekar/homecooked/internal/config"
	"github.com/pareshvernekar/homecooked/internal/database"
	"github.com/pareshvernekar/homecooked/internal/logger"
	"github.com/pareshvernekar/homecooked/internal/middleware"
)

func main() {
	// Initialize the configuration (Viper)
	config.Init()
	err := config.Read() // Read the configuration file
	if err != nil {
		logger.Logger.Error("Failed to read configuration", slog.Any("error", err))
		os.Exit(1)
	}
	// Get tenant ID from environment or use default
	tenantID := "1"
	if tenantIDStr := os.Getenv("TENANT_ID"); tenantIDStr != "" {
		if id, err := strconv.Atoi(tenantIDStr); err == nil {
			tenantID = strconv.Itoa(id)
		}
	}

	// Initialize the database connection with tenant isolation context
	if err := database.InitDB(tenantID); err != nil {
		logger.Logger.Error("Failed to initialize database", slog.Any("error", err))
		os.Exit(1)
	}

	// Ensure DB connection exists
	db := database.GetDB()
	if db == nil {
		logger.Logger.Error("Database connection is nil")
		os.Exit(1)
	}

	// Create the Gin router
	router := gin.Default()

	// Use the TenantMiddleware - extracts tenant ID from X-Tenant-ID header
	router.Use(middleware.TenantMiddleware())

	// Define a route that uses tenant isolation automatically
	router.GET("/", func(c *gin.Context) {
		tenantID, exists := c.Get(middleware.TenantIDKey)
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Tenant ID not found"})
			return
		}

		logger.Logger.Info("Handling request with tenant ID", slog.Any("tenant_id", tenantID))
		c.JSON(http.StatusOK, gin.H{"tenant_id": tenantID, "message": "Welcome to HomeCook!"})
	})

	router.GET("/orders", func(c *gin.Context) {
		tenantIDInt, exists := c.Get(middleware.TenantIDKey)
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Tenant ID not found"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		db := database.GetDB()
		if db == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection not available"})
			return
		}

		rows, err := db.QueryContext(ctx, "SELECT order_id, item_name FROM tenant_orders WHERE tenant_id = $1", tenantIDInt.(int))
		if err != nil {
			logger.Logger.Error("Failed to fetch orders", slog.Any("error", err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
			return
		}
		defer rows.Close()

		orders := []map[string]interface{}{}
		for rows.Next() {
			var orderID int
			var item_name string
			if err := rows.Scan(&orderID, &item_name); err != nil {
				logger.Logger.Error("Failed to parse orders", slog.Any("error", err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse orders"})
				return
			}
			orders = append(orders, map[string]interface{}{"id": orderID, "item_name": item_name})
		}

		logger.Logger.Info("Fetched orders for tenant", "tenant_id", tenantIDInt, "count", len(orders))
		c.JSON(http.StatusOK, gin.H{"orders": orders})
	})

	// Start the HTTP server
	if err := router.Run(":8080"); err != nil {
		logger.Logger.Error("Failed to start server", slog.Any("error", err))
		os.Exit(1)
	}

	// Note: The defer is for testing/graceful shutdown scenarios
	defer database.Close()
}
