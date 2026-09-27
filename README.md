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
# (starts Postgres via docker compose, applies schema on first boot, then HTTP API on :8080)
make local

# Build the Docker image, push it to the registry, and run the application
make deploy

# Clean up the build artifacts and Docker image
make clean

## Testing

### Prerequisites
- **Docker** (required for Postgres / Testcontainers)
- **Go** (same version as the project)

### Integration tests

Integration tests live under `tests/integration/` and exercise FoodCategory / FoodItem handlers against a real Postgres instance started via Testcontainers (Docker must be running). They do **not** require `make local`.

```bash
# Run all integration tests
go test ./tests/integration/... -count=1 -timeout 10m -v

# Or by package
go test ./tests/integration/foodcategories/ -count=1 -timeout 10m -v
go test ./tests/integration/fooditems/ -count=1 -timeout 10m -v
```

### Acceptance tests (Gherkin / Godog)

Acceptance tests live under `tests/acceptance/` and call the **running local HTTP API** (default `http://localhost:8080`). Start the server first, then run the suite in a second terminal.

```bash
# Terminal 1 — Postgres + schema + HTTP API
make local

# Terminal 2 — Gherkin acceptance scenarios
make acceptance-test
```

Equivalent direct command:

```bash
go test -tags=acceptance ./tests/acceptance/ -count=1 -timeout 5m -v
```

Optional environment overrides:

| Variable | Default | Purpose |
|----------|---------|---------|
| `ACCEPTANCE_BASE_URL` | `http://localhost:8080` | API base URL |
| `ACCEPTANCE_TENANT_ID` | `1` | Value sent as `X-Tenant-ID` |

If the schema is stale or missing tables after an older compose volume:

```bash
make db-reset   # wipe volume and re-apply scripts/init
make local
```

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