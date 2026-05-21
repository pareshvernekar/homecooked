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
	repo "github.com/pareshvernekar/homecooked/internal/repository"
	"github.com/pareshvernekar/homecooked/internal/server"
)

func main() {
	config.Init()
	err := config.Read()
	if err != nil {
		fmt.Printf("Error reading config: %v\n", err)
		os.Exit(1)
	}

	loggerInstance := logger.NewLogger()
	cacheClient, err := cache.NewCacheClient(cache.LoadConfig(config.Config), loggerInstance, nil)
	if err != nil {
		fmt.Printf("Warning: failed to initialize cache: %v\n", err)
	}

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

	var foodCategoryRepo repo.FoodCategoryRepository = repo.NewFoodCategoryRepository(db, tenantID)

	err = cacheClient.PostInitialize(context.Background(), tenantID, foodCategoryRepo)
	if err != nil {
		fmt.Printf("Warning: failed to initialize cache with food categories: %v\n", err)
	}

	foodItemRepo := repo.NewFoodItemRepository(db, tenantID)
	loggerInstance2 := logger.NewLogger()
	foodItemHandler := handlers.NewFoodItemHandler(foodItemRepo, loggerInstance2, cacheClient)

	srv := server.NewServer(db, loggerInstance, cacheClient)
	router := srv.Router
	gin.SetMode(gin.ReleaseMode)
	server.SetupRoutes(router, db, loggerInstance, foodItemHandler, cacheClient)

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
