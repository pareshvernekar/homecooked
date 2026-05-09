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

	cors "github.com/gin-contrib/cors"
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
	router := setupGinEngine()

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
	// Implementation for health checking
	return true
}

// GetVersion returns the current application version
func GetVersion() string {
	// Return the current version
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

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")

	v1 := api.Group("/")

	// Food Items Routes
	foodItems := v1.Group("/food-items")
	foodItems.GET("", handlers.NewFoodItemHandler(nil, nil).GetFoodItems)
	foodItems.POST("", handlers.NewFoodItemHandler(nil, nil).CreateFoodItem)
	foodItems.PUT("/:id", handlers.NewFoodItemHandler(nil, nil).UpdateFoodItem)
	foodItems.DELETE("/:id", handlers.NewFoodItemHandler(nil, nil).DeleteFoodItem)

	// Categories Routes (placeholder)
	categories := v1.Group("/categories")
	categories.POST("", func(c *gin.Context) {
		// Category creation logic
		c.Status(201)
	})

	// Weekly Menus Routes (placeholder)
	weeklyMenus := v1.Group("/weekly-menus")
	weeklyMenus.GET("", func(c *gin.Context) {
		// Get weekly menus
		c.Status(200)
	})

	// Catering Menus Routes (placeholder)
	cateringMenus := v1.Group("/catering-menus")
	cateringMenus.POST("", func(c *gin.Context) {
		// Create catering menu
		c.Status(201)
	})

	cateringMenus.GET("", func(c *gin.Context) {
		// Get catering menus
		c.Status(200)
	})

	// Menu Items within Menu (placeholder)
	v1.GET("/menus/:menuType/:menuId/menu-items", func(c *gin.Context) {
		// List menu items within a specific menu
		c.Status(200)
	})

	// Orders Routes (placeholder)
	v1.POST("/orders", func(c *gin.Context) {
		// Create order
		c.Status(201)
	})

	v1.GET("/orders", func(c *gin.Context) {
		// List orders
		c.Status(200)
	})

	// Notifications Routes (placeholder)
	v1.POST("/notifications", func(c *gin.Context) {
		// Create notification
		c.Status(201)
	})

	v1.GET("/notifications/:notificationId", func(c *gin.Context) {
		// Get notification by ID
		c.Status(200)
	})

	v1.DELETE("/notifications/:notificationId", func(c *gin.Context) {
		// Delete notification
		c.Status(204)
	})
}

// Helper function to initialize the Gin engine with global middleware and configuration
func setupGinEngine(logger *slog.Logger) *gin.Engine {

	// Create a new Gin engine with JSON logger
	gin.SetMode(gin.ReleaseMode)

	// Create a new Gin engine
	router := gin.New()

	// Security middleware
	router.Use(gin.Recovery())

	// Apply custom logger with JSON formatting using slog
	router.Use(jsonLogger(logger))
	// Enable CORS (Cross-Origin Resource Sharing)
	// Enable CORS (Cross-Origin Resource Sharing) with proper configuration
	router.Use(cors.Options{
		AllowAllOrigins:  []string{"*"}, // Allow all origins
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Tenant-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour, // Allow browsers to cache CORS preflight responses
	}.Handler())

	// Serve static assets if needed (optional)
	// router.Static("/static", "./public/static")

	return router
}

// jsonLogger wraps slog to provide JSON-formatted logging for Gin
func jsonLogger(slogger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Extract tenant ID from context if available
		tenantID := c.GetString(middleware.TenantIDKey)

		// Log request start with basic info
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
