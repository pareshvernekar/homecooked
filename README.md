# Homecooked Food Menu System\n\nA REST API for managing food catalogs, weekly menus, catering menus, orders, and notifications with multi-tenancy support.\n\n## Project Structure\n- cmd/: Main application entry point\n- internal/: Core application packages\n- migrations/: Database migration scripts\n- tests/: Test files\n- docs/: Documentation\n- config/: Configuration files\n- logs/: Application logs\n- pkg/: Third-party packages\n\n## Setup\n1. Install Go dependencies\n2. Set up environment variables\n3. Run migrations\n4. Start the server\n



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