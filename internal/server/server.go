package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/jmoiron/sqlx"
	handlers "github.com/pareshvernekar/homecooked/internal/handlers"
	logger "github.com/pareshvernekar/homecooked/internal/logger"
	middleware "github.com/pareshvernekar/homecooked/internal/middleware"
)

// Server represents the HTTP server structure
type Server struct {
	Router *gin.Engine
	DB     *sqlx.DB
	Logger *logger.Logger
}

// NewServer creates a new server instance with proper initialization
func NewServer(db *sqlx.DB, l *logger.Logger) *Server {
	router := setupGinEngine(l)

	return &Server{
		Router: router,
		DB:     db,
		Logger: l,
	}
}

// setupGinEngine configures the Gin router with middleware and routes
func setupGinEngine(l *logger.Logger) *gin.Engine {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.TenantMiddleware(l))

	return router
}

// SetupRoutes configures all HTTP routes for the server
func SetupRoutes(
	router *gin.Engine,
	_ *sqlx.DB,
	_ *logger.Logger,
	foodCategoryHandler *handlers.FoodCategoryHandler,
	foodItemHandler *handlers.FoodItemHandler,
	sizeUnitHandler *handlers.SizeUnitHandler,
	menuHandler *handlers.MenuHandler,
	orderHandler *handlers.OrderHandler,
) {
	v1 := router.Group("/api/v1")

	v1.GET("/categories", foodCategoryHandler.ListCategories)
	v1.POST("/categories", foodCategoryHandler.CreateCategory)
	v1.PUT("/categories/:id", foodCategoryHandler.UpdateCategory)
	v1.DELETE("/categories/:id", foodCategoryHandler.DeleteCategory)

	v1.GET("/food-items", foodItemHandler.GetFoodItems)
	v1.POST("/food-items", foodItemHandler.CreateFoodItem)
	v1.PUT("/food-items/:id", foodItemHandler.UpdateFoodItem)
	v1.DELETE("/food-items/:id", foodItemHandler.DeleteFoodItem)

	v1.GET("/size-units", sizeUnitHandler.ListSizeUnits)
	v1.POST("/size-units", sizeUnitHandler.CreateSizeUnit)
	v1.DELETE("/size-units/:id", sizeUnitHandler.DeleteSizeUnit)

	v1.POST("/menus", menuHandler.CreateMenu)
	v1.GET("/menus", menuHandler.ListMenus)
	v1.GET("/menus/:id", menuHandler.GetMenu)
	v1.PUT("/menus/:id", menuHandler.UpdateMenu)
	v1.DELETE("/menus/:id", menuHandler.DeleteMenu)
	v1.POST("/menus/:id/publish", menuHandler.PublishMenu)
	v1.POST("/menus/:id/unpublish", menuHandler.UnpublishMenu)

	v1.POST("/menus/:id/items", menuHandler.AddMenuItem)
	v1.DELETE("/menus/:id/items/:itemId", menuHandler.DeleteMenuItem)
	v1.POST("/menus/:id/items/:itemId/components/:componentId/size-options", menuHandler.AddSizeOption)
	v1.PUT("/menus/:id/items/:itemId/components/:componentId/size-options/:optionId", menuHandler.UpdateSizeOption)
	v1.DELETE("/menus/:id/items/:itemId/components/:componentId/size-options/:optionId", menuHandler.DeleteSizeOption)

	v1.POST("/orders", orderHandler.CreateOrder)
	v1.GET("/orders", orderHandler.ListOrders)
	v1.GET("/orders/:id", orderHandler.GetOrder)
	v1.PATCH("/orders/:id", orderHandler.UpdateOrder)
	v1.POST("/orders/:id/items", orderHandler.AddOrderItem)
	v1.PATCH("/orders/:id/items/:itemId", orderHandler.UpdateOrderItem)
	v1.DELETE("/orders/:id/items/:itemId", orderHandler.DeleteOrderItem)
	v1.POST("/orders/:id/payments", orderHandler.CreatePayment)
	v1.GET("/orders/:id/payments", orderHandler.ListPayments)
}

// Run starts the HTTP server and blocks until ctx is cancelled or Listen fails.
func (s *Server) Run(ctx context.Context) error {
	addr := ":8080"

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           s.Router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", addr, err)
	}

	s.Logger.Info(ctx, "HTTP server listening", "address", addr)
	fmt.Printf("HTTP server listening on http://localhost%s\n", addr)

	errCh := make(chan error, 1)
	go func() {
		err := httpServer.Serve(ln)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		s.Logger.Info(ctx, "Shutting down HTTP server")
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("http shutdown: %w", err)
		}
		return <-errCh
	case err := <-errCh:
		return err
	}
}
