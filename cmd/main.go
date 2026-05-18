package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/gin-gonic/gin"

	cache "github.com/pareshvernekar/homecooked/internal/cache"
	"github.com/pareshvernekar/homecooked/internal/config"
	"github.com/pareshvernekar/homecooked/internal/database"
	"github.com/pareshvernekar/homecooked/internal/handlers"
	logger "github.com/pareshvernekar/homecooked/internal/logger"
	"github.com/pareshvernekar/homecooked/internal/repository"
	"github.com/pareshvernekar/homecooked/internal/server"
)

func main() {
	// Initialize application configuration (Viper) - loads from config.yaml
	config.Init()
	err := config.Read() // Read the configuration file and bind to Viper
	if err != nil {
		fmt.Printf("Error reading config: %v\n", err)
		os.Exit(1)
	}

	// Initialize cache client using Viper config from main app
	// No YAML file needed for cache - all config comes through Viper
	cacheClient, err := cache.NewCacheClient(cache.LoadConfig(config.Config))
	if err != nil {
		fmt.Printf("Warning: failed to initialize cache: %v\n", err)
	}

	// Get tenant ID from environment or use default
	tenantID := "1"
	if tenantIDStr := os.Getenv("TENANT_ID"); tenantIDStr != "" {
		if id, err := strconv.Atoi(tenantIDStr); err == nil {
			tenantID = strconv.Itoa(id)
		}
	}

	// Initialize the logger
	loggerInstance := logger.NewLogger()

	// Initialize the database connection with tenant isolation context
	if err := database.InitDB(tenantID, loggerInstance); err != nil {
		fmt.Println("Failed to initialize database:", err)
		os.Exit(1)
	}

	// Ensure DB connection exists
	db := database.GetDB(loggerInstance)
	if db == nil {
		fmt.Println("Database connection is nil")
		os.Exit(1)
	}

	// Create FoodItemRepository using dependency injection
	foodItemRepo := repository.NewFoodItemRepository(db, tenantID)

	// Create handler with repository via dependency injection - separate logger for each handler
	loggerInstance2 := logger.NewLogger()
	foodItemHandler := handlers.NewFoodItemHandler(foodItemRepo, loggerInstance2, cacheClient)

	// Create server instance with repository injection and cache dependency
	srv := server.NewServer(db, loggerInstance, cacheClient)

	// Setup routes with handler that has repository dependency
	router := srv.Router
	gin.SetMode(gin.ReleaseMode)
	server.SetupRoutes(router, db, loggerInstance, foodItemHandler, cacheClient)

	// Create application context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		fmt.Println("Received shutdown signal...")
		cancel()
	}()

	// Start the server (in a goroutine)
	if err := srv.Run(ctx); err != nil {
		fmt.Printf("Server error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Service running successfully")
	select {} // Keep main alive until context is cancelled
}
