# Define the Go binary name
BINARY_NAME := homecooked

# Define the Go build flags
BUILD_FLAGS := -ldflags="-s -w"

# Define the Go test flags
TEST_FLAGS := -v

# Define the Go lint tool
LINT_TOOL := golangci-lint

# Define the Go migrate flags
MIGRATE_FLAGS := -config config/config.yaml

# Define the Docker image name
DOCKER_IMAGE := homecooked:latest

# Define the Docker build context
DOCKER_CONTEXT := ./cmd

# Define the local development port
DEV_PORT := 8080

# Define the Docker run command
DOCKER_RUN := docker run --rm -p 8080:8080 $(DOCKER_IMAGE)

# Define the Docker build command
DOCKER_BUILD := docker build -t $(DOCKER_IMAGE) $(DOCKER_CONTEXT)

# Define the Docker push command
DOCKER_PUSH := docker push $(DOCKER_IMAGE)


# Default target (runs test)
all: build


# ============================================
# Linting & Code Quality
# ============================================
lint: ## Run linters (golangci-lint)
	@echo "🔍 Running golangci-lint..."
	@which $(LINT_TOOL) || (go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest && echo "✅ Installed golangci-lint")
	@$(LINT_TOOL) run --timeout=5m ./...
	@echo "✅ Linting passed"


# ============================================
# Testing
# ============================================

# Define the test target
test: lint ## Run tests (includes linting check)
	@echo "Running tests..."
	@go test $(TEST_FLAGS) -cover -race ./...

# Define the unit-test target
unit-test: ## Run unit tests only
	@echo "🧪 Running unit tests..."
	@go test -v -cover ./internal/...

integration-test: ## Run integration tests with database (requires PostgreSQL running)
	@echo "🔗 Running integration tests..."
	@if ! docker ps | grep -q postgres; then \
		echo "❌ No PostgreSQL container found. Start one with: docker run --name pg -e POSTGRES_DB=homecooked -p 5432:5432 -d postgres"; \
		exit 1; \
	fi
	go test $(TEST_FLAGS) -tags=integration ./...

acceptance-test: ## Run Gherkin acceptance tests against a running local server
	@echo "✅ Running acceptance tests against $${ACCEPTANCE_BASE_URL:-http://localhost:8080}..."
	@echo "   (Start the server first with: make local)"
	go test $(TEST_FLAGS) -tags=acceptance ./tests/acceptance/ -count=1 -timeout 5m


# ============================================
# Building & Running
# ============================================

# Define the build target
build: ## Build Go binary
	@echo "Building the Go binary..."
	@go build $(BUILD_FLAGS) -o $(BINARY_NAME) $(DOCKER_CONTEXT)
	@echo "✅ Binary built successfully"
	@ls -lh $(BINARY_NAME)

db-up: ## Start Postgres via docker compose and wait until ready
	@echo "🐘 Starting Postgres (docker compose)..."
	@docker compose up -d db
	@echo "⏳ Waiting for Postgres to accept connections..."
	@for i in $$(seq 1 30); do \
		if docker compose exec -T db pg_isready -U postgres -d homecooked >/dev/null 2>&1; then \
			echo "✅ Postgres is ready (db=homecooked)"; \
			exit 0; \
		fi; \
		sleep 1; \
	done; \
	echo "❌ Postgres did not become ready in time"; \
	docker compose logs db | tail -40; \
	exit 1

db-down: ## Stop compose database (keeps volume/data)
	@docker compose down

db-reset: ## Recreate Postgres volume and re-apply init schema
	@echo "♻️  Recreating Postgres with fresh schema from scripts/init..."
	@docker compose down -v
	@$(MAKE) db-up

# Define the local development target: DB + schema init + HTTP server
local: db-up ## Start Postgres (schema init on first boot) + HTTP API on :8080
	@echo "🚀 Starting HTTP server against localhost:5432/homecooked..."
	@echo "   Schema source: scripts/init/01_food_catalog.sql (applied on first volume init)"
	@echo "   Reset schema with: make db-reset"
	@go run ./cmd/main.go

dev: ## Run app in development mode with structured logging and debug output (no hot-reload for now)
	@echo "🚀 Starting application in dev mode..."
	@$(MAKE) db-up
	go run ./cmd/main.go &\
	(PID=$$!; echo "⏳ Application running with PID $$PID"; sleep infinity); \
	exit 0

# Define the run target
run: ## Run built binary (requires make build first)
	@echo "▶️ Running application..."
	@if [ ! -f $(BINARY_NAME) ]; then \
		echo "❌ Binary not found. Run 'make build' first."; \
		exit 1; \
	fi
	@if [ ! -x $(BINARY_NAME) ]; then chmod +x $(BINARY_NAME); fi;
	@./$(BINARY_NAME)


server: local ## Alias for make local (DB + HTTP server)



# ============================================
# Database Migrations
# ============================================

# Define the migrate target
migrate:
	@echo "Running database migrations..."
	@go run github.com/golang-migrate/migrate/v4@v4.14.1 $(MIGRATE_FLAGS) up

migrate-up: ## Run database migrations forward
	@echo "🔄 Running migrations UP..."
	@golang-migrate -path internal/config/config.yaml --database postgresql://localhost:5432/homecooked up
	@echo "✅ Migrations completed"

migrate-down: ## Rollback one migration
	@echo "⏮️ Rolling back migrations..."
	@golang-migrate -path internal/config/config.yaml --database postgresql://localhost:5432/homecooked down 1
	@echo "✅ Migration rolled back"

migrate-status: ## Show migration status
	@echo "📊 Migration status:"
	@golang-migrate -path internal/config/config.yaml --database postgresql://localhost:5432/homecooked status



# ============================================
# Docker Targets
# ============================================
docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	docker build -t $(DOCKER_IMAGE) $(DOCKER_CONTEXT)
	@echo "✅ Image built successfully"

docker-run: ## Run application in Docker
	@echo "🚀 Running in Docker..."
	docker run --rm -p $(DEV_PORT):8080 $(DOCKER_IMAGE)

docker-stop: ## Stop running container
	docker stop homecooked-container 2>/dev/null || true


# Define the deploy target
deploy:
	@echo "Building the Docker image..."
	@$(DOCKER_BUILD)
	@echo "Pushing the Docker image to the registry..."
	@$(DOCKER_PUSH)
	@echo "Deploying the application..."
	@$(DOCKER_RUN)

# ============================================
# Cleanup & Maintenance
# ============================================

clean: ## Remove build artifacts and caches
	@echo "🧹 Cleaning up..."
	rm -f $(BINARY_NAME)
	go clean -cache -testcache -i
	docker rmi $(DOCKER_IMAGE) 2>/dev/null || true
	@echo "✅ Cleaned successfully"

help: ## Show this help message
	@grep -E '^\.PHONY|^##?' $$(grep -E '^\w+:' Makefile | awk -F: '{print $$1}')" | \
	  sed -n 'p' | head -20 || true
	@echo ""
	@echo "📖 HomeCooked Makefile Help"
	@echo ""
	@echo "Development:"
	@echo "  make local     Start Postgres + schema + HTTP API (:8080)"
	@echo "  make db-up     Start Postgres only (compose)"
	@echo "  make db-reset  Wipe volume and re-apply scripts/init schema"
	@echo "  make db-down   Stop Postgres compose services"
	@echo "  make dev       Run app in development mode (with logs)"
	@echo "  make server    Alias for make local"
	@echo "  make run       Run compiled binary"
	@echo ""
	@echo "Testing:"
	@echo "  make test      Run all tests (includes linting)"
	@echo "  make unit-test Run unit tests only"
	@echo "  make integration-test Run integration tests (requires PostgreSQL)"
	@echo "  make acceptance-test Run Gherkin API acceptance tests"
	@echo ""
	@echo "Building:"
	@echo "  make build         Build Go binary"
	@echo "  make docker-build  Build Docker image"
	@echo "  make docker-run    Run in Docker"
	@echo ""
	@echo "Database:"
	@echo "  make migrate-up    Apply all migrations forward"
	@echo "  make migrate-down  Rollback one migration"
	@echo "  make migrate-status Show current migration status"
	@echo ""
	@echo "Maintenance:"
	@echo "  make clean   Clean caches and build artifacts"