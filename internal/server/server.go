package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/jmoiron/sqlx"
	"github.com/pareshvernekar/homecooked/internal/config"
	"github.com/pareshvernekar/homecooked/internal/handlers"
	"github.com/pareshvernekar/homecooked/internal/middleware"
)

// Server represents the HTTP server structure
type Server struct {
	Router *gin.Engine
	DB     *sqlx.DB
	Logger *slog.Logger
}

// NewServer creates a new server instance with proper initialization
func NewServer(db *sqlx.DB, logger *slog.Logger) *Server {
	router := setupGinEngine(logger)

	return &Server{
		Router: router,
		DB:     db,
		Logger: logger,
	}
}

// Run starts the server and handles graceful shutdown
func (s *Server) Run(ctx context.Context) error {
	port := config.Config.GetInt("server.port")

	// Apply the global middleware to all routes
	s.Router.Use(middleware.TenantMiddleware())

	addr := fmt.Sprintf(":%d", port)
	s.Logger.Info("Starting server", "address", addr)

	server := &http.Server{
		Addr:         addr,
		Handler:      s.Router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start the server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			s.Logger.Error("Server failed to start", "error", err)
			return
		}
	}()

	// Wait for interrupt signal (SIGINT, SIGTERM)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	s.Logger.Info("Server shutdown initiated...")

	// Graceful shutdown
	if err := s.gracefulShutdown(server); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	s.Logger.Info("Server stopped gracefully")
	return nil
}

// Health check endpoint
func HealthCheck() bool {
	return true
}

// GetVersion returns the current application version
func GetVersion() string {
	return "1.0.0"
}

// gracefulShutdown performs a graceful server shutdown
func (s *Server) gracefulShutdown(server *http.Server) error {
	// Create context with 30 second timeout for graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.Logger.Info("Initiating graceful shutdown...")

	// Shutdown the HTTP server gracefully
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	s.Logger.Info("Server shutdown completed successfully")
	return nil
}

// SetupRoutes configures all API routes with dependency injection
func SetupRoutes(router *gin.Engine, db *sqlx.DB, logger *slog.Logger, foodItemHandler *handlers.FoodItemHandler) {
	api := router.Group("/api/v1")

	v1 := api.Group("/")

	// Food Items Routes - Using Dependency Injection
	foodItems := v1.Group("/food-items")
	foodItems.GET("", foodItemHandler.GetFoodItems)
	foodItems.POST("", foodItemHandler.CreateFoodItem)
	foodItems.PUT("/:id", foodItemHandler.UpdateFoodItem)
	foodItems.DELETE("/:id", foodItemHandler.DeleteFoodItem)

	// Categories Routes (placeholder)
	categories := v1.Group("/categories")
	categories.POST("", func(c *gin.Context) {
		c.Status(201)
	})

	// Weekly Menus Routes (placeholder)
	weeklyMenus := v1.Group("/weekly-menus")
	weeklyMenus.GET("", func(c *gin.Context) {
		c.Status(200)
	})

	// Catering Menus Routes (placeholder)
	cateringMenus := v1.Group("/catering-menus")
	cateringMenus.POST("", func(c *gin.Context) {
		c.Status(201)
	})

	cateringMenus.GET("", func(c *gin.Context) {
		c.Status(200)
	})

	// Menu Items within Menu (placeholder)
	v1.GET("/menus/:menuType/:menuId/menu-items", func(c *gin.Context) {
		c.Status(200)
	})

	// Orders Routes (placeholder)
	v1.POST("/orders", func(c *gin.Context) {
		c.Status(201)
	})

	v1.GET("/orders", func(c *gin.Context) {
		c.Status(200)
	})

	// Notifications Routes (placeholder)
	v1.POST("/notifications", func(c *gin.Context) {
		c.Status(201)
	})

	v1.GET("/notifications/:notificationId", func(c *gin.Context) {
		c.Status(200)
	})

	v1.DELETE("/notifications/:notificationId", func(c *gin.Context) {
		c.Status(204)
	})
}

// Helper function to initialize the Gin engine with global middleware and configuration
func setupGinEngine(logger *slog.Logger) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Security middleware
	router.Use(gin.Recovery())

	// Apply custom logger with JSON formatting using slog
	router.Use(jsonLogger(logger))

	return router
}

// jsonLogger wraps slog to provide JSON-formatted logging for Gin
func jsonLogger(slogger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Extract tenant ID from context if available
		tenantID := c.GetString(middleware.TenantIDKey)

		c.Next()

		// Log response with timing and status
		duration := time.Since(start)
		statusCode := c.Writer.Status()

		slogger.Log(
			c.Request.Context(),
			slog.LevelInfo,
			fmt.Sprintf("HTTP %s %s - %d - %.2fs",
				c.Request.Method,
				c.Request.URL.Path,
				statusCode,
				duration.Seconds()),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status_code", statusCode),
			slog.Duration("duration", duration),
			slog.String("remote_addr", c.ClientIP()),
		)

		if tenantID != "" {
			slogger.Log(c.Request.Context(), slog.LevelInfo,
				fmt.Sprintf("Tenant %s accessed endpoint: %s", tenantID, c.Request.URL.Path),
				slog.String("tenant_id", tenantID),
				slog.String("path", c.Request.URL.Path),
			)
		}
	}
}
