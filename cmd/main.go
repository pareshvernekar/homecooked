package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	cache "github.com/pareshvernekar/homecooked/internal/cache"
	"github.com/pareshvernekar/homecooked/internal/config"
	models "github.com/pareshvernekar/homecooked/internal/models"
	"github.com/pareshvernekar/homecooked/internal/database"
	"github.com/pareshvernekar/homecooked/internal/handlers"
	logger "github.com/pareshvernekar/homecooked/internal/logger"
	repo "github.com/pareshvernekar/homecooked/internal/repository"
	"github.com/pareshvernekar/homecooked/internal/server"
)

// =============================================================================
// CACHE CONFIGURATION TYPES
// These types satisfy the internal cache package's CacheConfig interface.
// =============================================================================

// loadCacheConfig creates a cache configuration for food category caching.
func loadCacheConfig() cache.CacheConfig {
	return cache.CacheConfig{
		DefaultTTL:                 30 * time.Minute,
		MaxItems:          1000,
		TTLOverrides: map[string]time.Duration{
			"food_category":    30 * time.Minute,
			"food_item":        15 * time.Minute,
			"weekly_menu":      60 * time.Minute,
			"catering_menu":    60 * time.Minute,
			"menu_items":       15 * time.Minute,
			"order_summary":    30 * time.Minute,
			"catering_order":   60 * time.Minute,
		},
	}
}

func main() {
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

	cacheConfig := loadCacheConfig()

	// =============================================================================
	// Initialize ALL repositories for multi-entity caching
	// =============================================================================
	// Create a map of entity_type -> repository so all caches can access the same repos.
	// This is how GenericCacheClient[T] stores repos internally and iterates over them
	// during PostInitialize to populate the cache with ALL entity types.
	allRepos := make(map[string]any)

	var foodCategoryRepo repo.FoodCategoryRepository = repo.NewFoodCategoryRepository(db, tenantID)
	allRepos["food_category"] = foodCategoryRepo

	var foodItemRepo repo.FoodItemRepository = repo.NewFoodItemRepository(db, tenantID)
	allRepos["food_item"] = foodItemRepo

	// =============================================================================
	// Initialize Generic Cache Clients with ALL repositories
	// =============================================================================
	// Each entity type gets its own GenericCacheClient instance. All clients store
	// allRepos in their repos field so PostInitialize can iterate over all caches.
	//
	// NOTE: The generic type parameter T is now the MODEL type, not the repository:
	//   - models.FoodCategory (not repo.PostgreSQLFoodCategoryRepository)
	//   - models.FoodItem (not repo.PostgreSQLFoodItemRepository)
	// This allows Set() to accept model values directly.
	var foodCategoryCache cache.TypedClient[models.FoodCategory] =
		cache.NewGenericFoodCategoryCache(cacheConfig, loggerInstance)

	var foodItemCache cache.TypedClient[models.FoodItem] =
		cache.NewGenericFoodItemCache(cacheConfig, loggerInstance)

	// Initialize all caches with ALL repositories at once. Each client iterates
	// through its shared repos map and populates its store. This enables type-safe
	// access to any cached entity using runtime type assertions.
	if err := foodCategoryCache.PostInitialize(context.Background(), tenantID, allRepos); err != nil {
		loggerInstance.Error(context.Background(), "Failed to initialize food category cache", "error", err)
	}
	if err := foodItemCache.PostInitialize(context.Background(), tenantID, allRepos); err != nil {
		loggerInstance.Error(context.Background(), "Failed to initialize food item cache", "error", err)
	}

	// =============================================================================
	// Create handlers and server with proper cache injection
	// =============================================================================
	foodItemHandler := handlers.NewFoodItemHandler(foodItemRepo, loggerInstance, foodItemCache)

	srv := server.NewServer(db, loggerInstance, foodItemCache)
	router := srv.Router
	gin.SetMode(gin.ReleaseMode)
	server.SetupRoutes(router, db, loggerInstance, foodItemHandler, foodItemCache)

	ctx, cancel := context.WithCancel(context.Background())
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
