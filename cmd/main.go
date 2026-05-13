package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/pareshvernekar/homecooked/internal/config"
	"github.com/pareshvernekar/homecooked/internal/database"
	"github.com/pareshvernekar/homecooked/internal/handlers"
	"github.com/pareshvernekar/homecooked/internal/logger"
	"github.com/pareshvernekar/homecooked/internal/repository"
	"github.com/pareshvernekar/homecooked/internal/server"
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

	// Create FoodItemRepository using dependency injection
	// Repository gets access to database through constructor, not direct import
	foodItemRepo := repository.NewFoodItemRepository(db, tenantID)

	// Create handler with repository via dependency injection
	foodItemHandler := handlers.NewFoodItemHandler(foodItemRepo)

	// Create server instance with repository injection
	srv := server.NewServer(db, logger.Logger)

	// Setup routes with handler that has repository dependency
	server.SetupRoutes(srv.Router, db, logger.Logger, foodItemHandler)

	// Create application context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("⏳ Received shutdown signal...")
		cancel()
	}()

	// Start the server (in a goroutine)
	if err := srv.Run(ctx); err != nil {
		log.Fatalf("❌ Server error: %v", err)
		os.Exit(1)
	}

	log.Println("✅ Service running successfully")
	select {} // Keep main alive until context is cancelled
}
