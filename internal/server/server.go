package server

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/jmoiron/sqlx"
	cache "github.com/pareshvernekar/homecooked/internal/cache"
	handlers "github.com/pareshvernekar/homecooked/internal/handlers"
	logger "github.com/pareshvernekar/homecooked/internal/logger"
	middleware "github.com/pareshvernekar/homecooked/internal/middleware"
	models "github.com/pareshvernekar/homecooked/internal/models"
)

// Server represents the HTTP server structure
type Server struct {
	Router *gin.Engine
	DB     *sqlx.DB
	Logger *logger.Logger
	Cache  cache.TypedClient[models.FoodItem] // Cache client for food item caching (generic type)
}

// NewServer creates a new server instance with proper initialization
func NewServer(db *sqlx.DB, l *logger.Logger, c cache.TypedClient[models.FoodItem]) *Server {
	router := setupGinEngine(l)

	return &Server{
		Router: router,
		DB:     db,
		Logger: l,
		Cache:  c,
	}
}

// setupGinEngine configures the Gin router with middleware and routes
func setupGinEngine(l *logger.Logger) *gin.Engine {
	router := gin.New()

	// Setup middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	return router
}

// SetupRoutes configures all HTTP routes for the server
func SetupRoutes(
	router *gin.Engine,
	db *sqlx.DB,
	l *logger.Logger,
	handler *handlers.FoodItemHandler,
	cache cache.TypedClient[models.FoodItem],
) {
	// Food items API
	v1 := router.Group("/api/v1")
	{
		v1.GET("/food-items", handler.GetFoodItems)
		v1.POST("/food-items", handler.CreateFoodItem)
		v1.PUT("/food-items/:id", handler.UpdateFoodItem)
		v1.DELETE("/food-items/:id", handler.DeleteFoodItem)
	}

	// Initialize tenant ID in context (for middleware that needs it)
	router.Use(func(c *gin.Context) {
		c.Set(middleware.TenantIDKey, "default")
		c.Next()
	})
}

// Run starts the HTTP server
func (s *Server) Run(ctx context.Context) error {
	port := ":8080"

	addr := port
	s.Logger.Info(ctx, "Starting server", "address", addr)

	return s.Router.Run(addr)
}
