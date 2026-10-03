package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/pareshvernekar/homecooked/internal/config"
	"github.com/pareshvernekar/homecooked/internal/database"
	handlers "github.com/pareshvernekar/homecooked/internal/handlers"
	logger "github.com/pareshvernekar/homecooked/internal/logger"
	repo "github.com/pareshvernekar/homecooked/internal/repository"
	"github.com/pareshvernekar/homecooked/internal/server"
	foodcategoryService "github.com/pareshvernekar/homecooked/internal/services/foodcategory"
	fooditemService "github.com/pareshvernekar/homecooked/internal/services/fooditem"
	menuService "github.com/pareshvernekar/homecooked/internal/services/menu"
	menuitemService "github.com/pareshvernekar/homecooked/internal/services/menuitem"
	notificationService "github.com/pareshvernekar/homecooked/internal/services/notification"
	orderService "github.com/pareshvernekar/homecooked/internal/services/order"
	sizeunitService "github.com/pareshvernekar/homecooked/internal/services/sizeunit"
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
	var sizeUnitRepo = repo.NewSizeUnitRepository(db, loggerInstance, tenantID)
	var menuRepo = repo.NewMenuRepository(db, loggerInstance, tenantID)
	var menuItemRepo = repo.NewMenuItemRepository(db, loggerInstance, tenantID)
	var orderRepo = repo.NewOrderRepository(db, loggerInstance, tenantID)
	var notificationRepo = repo.NewNotificationRepository(db, loggerInstance)

	foodCategoryService := foodcategoryService.NewFoodCategoryService(foodCategoryRepo, loggerInstance)
	foodItemService := fooditemService.NewFoodItemService(foodItemRepo, loggerInstance, foodCategoryService)
	sizeUnitService := sizeunitService.NewService(sizeUnitRepo, loggerInstance)
	menuSvc := menuService.NewService(menuRepo, menuItemRepo, loggerInstance)
	menuItemSvc := menuitemService.NewService(menuItemRepo, menuSvc, sizeUnitService, loggerInstance)
	outboxBuilder := notificationService.NewBuilder(notificationRepo, loggerInstance)
	orderSvc := orderService.NewService(orderRepo, menuSvc, loggerInstance, orderService.WithNotifications(outboxBuilder))
	notificationSvc := notificationService.NewService(notificationRepo)

	foodCategoryHandler := handlers.NewFoodCategoryHandler(foodCategoryService, loggerInstance)
	foodItemHandler := handlers.NewFoodItemHandler(foodItemService, loggerInstance)
	sizeUnitHandler := handlers.NewSizeUnitHandler(sizeUnitService, loggerInstance)
	menuHandler := handlers.NewMenuHandler(menuSvc, menuItemSvc, loggerInstance)
	orderHandler := handlers.NewOrderHandler(orderSvc, loggerInstance)
	notificationHandler := handlers.NewNotificationHandler(notificationSvc, loggerInstance)

	srv := server.NewServer(db, loggerInstance)
	router := srv.Router
	gin.SetMode(gin.ReleaseMode)
	server.SetupRoutes(router, db, loggerInstance, foodCategoryHandler, foodItemHandler, sizeUnitHandler, menuHandler, orderHandler, notificationHandler)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var workerDone chan struct{}
	// REQNOTIF003, REQNOTIF004: async delivery worker (default on; NOTIFICATION_WORKER=0 disables).
	if notificationWorkerEnabled() {
		provider, err := notificationService.NewProvider(os.Getenv("SMS_PROVIDER"), notificationRepo, loggerInstance)
		if err != nil {
			fmt.Printf("Invalid SMS provider configuration: %v\n", err)
			os.Exit(1)
		}
		worker := notificationService.NewWorker(notificationRepo, provider, loggerInstance,
			notificationService.WorkerConfig{PollInterval: notificationPollInterval()})
		workerDone = make(chan struct{})
		go func() {
			defer close(workerDone)
			worker.Run(ctx)
		}()
	}

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

	cancel() // stop the worker before closing the DB
	if workerDone != nil {
		<-workerDone
	}

	if err := database.Close(loggerInstance); err != nil {
		fmt.Printf("Error closing database: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Server stopped cleanly")
}

// notificationWorkerEnabled reports whether the delivery worker should start (default on).
func notificationWorkerEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("NOTIFICATION_WORKER"))) {
	case "0", "false", "off", "no":
		return false
	default:
		return true
	}
}

// notificationPollInterval reads NOTIFICATION_POLL_INTERVAL (e.g. "5s"); zero uses the default.
func notificationPollInterval() time.Duration {
	if d, err := time.ParseDuration(os.Getenv("NOTIFICATION_POLL_INTERVAL")); err == nil && d > 0 {
		return d
	}
	return 0
}
