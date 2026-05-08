# Define the Go binary name
BINARY_NAME := homecooked

# Define the Go build flags
BUILD_FLAGS := -ldflags="-s -w"

# Define the Go test flags
TEST_FLAGS := -v

# Define the Go migrate flags
MIGRATE_FLAGS := -config config/config.yaml

# Define the Docker image name
DOCKER_IMAGE := homecooked:latest

# Define the Docker build context
DOCKER_CONTEXT := ./cmd

# Define the Docker run command
DOCKER_RUN := docker run --rm -p 8080:8080 $(DOCKER_IMAGE)

# Define the Docker build command
DOCKER_BUILD := docker build -t $(DOCKER_IMAGE) $(DOCKER_CONTEXT)

# Define the Docker push command
DOCKER_PUSH := docker push $(DOCKER_IMAGE)

# Define the build target
build:
	@echo "Building the Go binary..."
	@go build $(BUILD_FLAGS) -o $(BINARY_NAME) $(DOCKER_CONTEXT)

# Define the run target
run:
	@echo "Running the Go binary..."
	@./$(BINARY_NAME)

# Define the test target
test:
	@echo "Running tests..."
	@go test $(TEST_FLAGS) ./...

# Define the migrate target
migrate:
	@echo "Running database migrations..."
	@go run github.com/golang-migrate/migrate/v4@v4.14.1 $(MIGRATE_FLAGS) up

# Define the local development target
local:
	@echo "Running the application in local development mode..."
	@go run main.go

# Define the deploy target
deploy:
	@echo "Building the Docker image..."
	@$(DOCKER_BUILD)
	@echo "Pushing the Docker image to the registry..."
	@$(DOCKER_PUSH)
	@echo "Deploying the application..."
	@$(DOCKER_RUN)

# Define the clean target
clean:
	@echo "Cleaning up..."
	@rm -f $(BINARY_NAME)
	@docker rmi $(DOCKER_IMAGE) || true