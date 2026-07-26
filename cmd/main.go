package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/gin-gonic/gin"

	"github.com/pareshvernekar/homecooked/internal/config"
	"github.com/pareshvernekar/homecooked/internal/database"
	"github.com/pareshvernekar/homecooked/internal/handlers"
	logger "github.com/pareshvernekar/homecooked/internal/logger"
	models "github.com/pareshvernekar/homecooked/internal/models"
	repo "github.com/pareshvernekar/homecooked/internal/repository"
	"github.com/pareshvernekar/homecooked/internal/server"
	foodCategoryService "github.com/pareshvernekar/homecooked/internal/services/foodcategory"
	foodItemService "github.com/pareshvernekar/homecooked/internal/services/fooditem"
)

type FoodCategoryRepository interface {
	ListByTenant(ctx context.Context, tenantID string) ([]*models.FoodCategory, error)
	GetByID(ctx context.Context, tenantID string, id string) (*models.FoodCategory, error)
	Create(ctx context.Context, category *models.FoodCategory) error
	Update(ctx context.Context, category *models.FoodCategory) (int64, error)
	Delete(ctx context.Context, tenantID string, id string) (int64, error)
}

type FoodItemRepository interface {
	ListByTenant(ctx context.Context, tenantID string, offset, limit int) ([]*models.FoodItem, int64, error)
	GetByID(ctx context.Context, id string, tenantID string) (*models.FoodItem, error)
	Create(ctx context.Context, item *models.FoodItem) error
	Update(ctx context.Context, item *models.FoodItem) error
	Delete(ctx context.Context, id string) (int64, error)
}

type FoodItemService interface {
	List(ctx context.Context, tenantID string, page int, limit int) ([]*models.FoodItem, error)
	GetByID(ctx context.Context, id string, tenantID string) (*models.FoodItem, error)
	Create(ctx context.Context, createReq *models.FoodItemCreateRequest, tenantID string) (*models.FoodItem, error)
	Update(ctx context.Context, id string, updateReq *models.FoodItemUpdateRequest, tenantID string) error
	Delete(ctx context.Context, id string, tenantID string) (int64, error)
}

type FoodCategoryService interface {
	ListCategories(ctx context.Context, tenantID string) ([]*models.FoodCategory, error)
	GetCategoryByID(ctx context.Context, tenantID string, id string) (*models.FoodCategory, error)
	CreateCategory(ctx context.Context, category *models.FoodCategory) error
	UpdateCategory(ctx context.Context, category *models.FoodCategory) (int64, error)
	DeleteCategory(ctx context.Context, tenantID string, id string) (int64, error)
}

func main() {
	ctx := context.Background()
	config.Init()

	err := config.Read()
	if err != nil {
		fmt.Printf("Error reading config: %v\n", err)
		os.Exit(1)
	}

	loggerInstance := logger.NewLogger()

	tenantID := "1"
	if tenantIDStr := os.Getenv("TENANT_ID"); tenantIDStr != "" {
		if id, err := strconv.Atoi(tenantIDStr); err == nil {
			tenantID = strconv.Itoa(id)
		}
	}

	if err := database.InitDB(tenantID, loggerInstance); err != nil {
		fmt.Println("Failed to initialize database:", err)
		os.Exit(1)
	}

	db := database.GetDB(loggerInstance)
	if db == nil {
		fmt.Println("Database connection is nil")
		os.Exit(1)
	}

	// =============================================================================
	// Initialize ALL repositories for multi-entity caching
	// =============================================================================
	// Create a map of entity_type -> repository so all caches can access the same repos.
	// This is how GenericCacheClient[T] stores repos internally and iterates over them
	// during PostInitialize to populate the cache with ALL entity types.

	var foodCategoryRepo FoodCategoryRepository = repo.NewFoodCategoryRepository(db, loggerInstance, tenantID)

	var foodItemRepo FoodItemRepository = repo.NewFoodItemRepository(db, loggerInstance, tenantID)

	foodCategoryService := foodCategoryService.NewFoodCategoryService(foodCategoryRepo, loggerInstance)
	foodItemService := foodItemService.NewFoodItemService(foodItemRepo, loggerInstance, foodCategoryService)

	// =============================================================================
	// Create handlers and server with proper cache injection
	// =============================================================================
	foodItemHandler := handlers.NewFoodItemHandler(foodItemService, loggerInstance)

	srv := server.NewServer(db, loggerInstance)
	router := srv.Router
	gin.SetMode(gin.ReleaseMode)
	server.SetupRoutes(router, db, loggerInstance, foodItemHandler)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		fmt.Println("Received shutdown signal...")
		cancel()
	}()

	if err := srv.Run(ctx); err != nil {
		fmt.Printf("Server error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Service running successfully")
	select {}
}
