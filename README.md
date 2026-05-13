# Homecooked Food Menu System\n\nA REST API for managing food catalogs, weekly menus, catering menus, orders, and notifications with multi-tenancy support.\n\n## Project Structure\n- cmd/: Main application entry point\n- internal/: Core application packages\n- migrations/: Database migration scripts\n- tests/: Test files\n- docs/: Documentation\n- config/: Configuration files\n- logs/: Application logs\n- pkg/: Third-party packages\n\n## Setup\n1. Install Go dependencies\n2. Set up environment variables\n3. Run migrations\n4. Start the server\n


#Install golangci-lint (if not already installed):
```
brew install golangci-lint  # For macOS
```

OR

```
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

# Docker required to run PostGreSQL

## Start the Postgres

### To create the DB
```
docker compose -f docker-compose.yml up -d db
```

```
 docker compose up -d
```

## Stop the Postgres

 ```
 docker compose down
```

## Stop Postgres and clean the volume (Delete the tables)
```
docker compose down -v
```


# Make commands

# Build the Go binary
make build

# Run the Go binary
make run

# Run tests
make test

# Run database migrations
make migrate

# Run the application in local development mode
make local

# Build the Docker image, push it to the registry, and run the application
make deploy

# Clean up the build artifacts and Docker image
make clean

# Docker commands

 ## How to Run

  # Build, start DB, and app together
  docker compose -f docker-compose.prod.yml up --build -d

  # Wait for initialization (DB schema loads from init.sql)
  sleep 20

  # Verify both containers are healthy
  docker compose ps

  # Access the API
  curl http://localhost:8080/health

  # Stop everything
  docker compose -f docker-compose.prod.yml down

  # Restart without removing data
  docker compose -f docker-compose.prod.yml up -d