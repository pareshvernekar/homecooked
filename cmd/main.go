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
	handlers "github.com/pareshvernekar/homecooked/internal/handlers"
	logger "github.com/pareshvernekar/homecooked/internal/logger"
	repo "github.com/pareshvernekar/homecooked/internal/repository"
	"github.com/pareshvernekar/homecooked/internal/server"
	foodcategoryService "github.com/pareshvernekar/homecooked/internal/services/foodcategory"
	fooditemService "github.com/pareshvernekar/homecooked/internal/services/fooditem"
)

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

	var foodCategoryRepo = repo.NewFoodCategoryRepository(db, loggerInstance, tenantID)
	var foodItemRepo = repo.NewFoodItemRepository(db, loggerInstance, tenantID)

	foodCategoryService := foodcategoryService.NewFoodCategoryService(foodCategoryRepo, loggerInstance)
	foodItemService := fooditemService.NewFoodItemService(foodItemRepo, loggerInstance, foodCategoryService)

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
