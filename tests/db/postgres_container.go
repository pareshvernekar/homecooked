package db

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PostgresContainer represents a PostgreSQL database instance managed by testcontainers-go
type PostgresContainer struct {
	container testcontainers.Container
	host      string
	stopFunc  func() error
}

// NewPostgresContainer creates a new PostgreSQL database container using TestContainers
func NewPostgresContainer(ctx context.Context) (*PostgresContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:14-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env:          map[string]string{"POSTGRES_USER": "user", "POSTGRES_PASSWORD": "secretpassword123"},
		WaitingFor: wait.ForAll( // Wait for log message twice (startup + readiness)
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(10*time.Second), wait.ForListeningPort("5432/tcp"),
		),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start PostgreSQL container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get host: %w", err)
	}

	stopFunc := func() error {
		return container.Terminate(ctx)
	}

	return &PostgresContainer{
		container: container,
		host:      host,
		stopFunc:  stopFunc,
	}, nil
}

// Terminate terminates the PostgreSQL container
func (pc *PostgresContainer) Terminate(ctx context.Context) error {
	return pc.container.Terminate(ctx)
}

// Stop stops the container without termination
func (pc *PostgresContainer) Stop(ctx context.Context) error {
	// Grace period before Docker forces a kill (e.g., 5 seconds)
	stopTimeout := 5 * time.Second
	return pc.container.Stop(ctx, &stopTimeout)
}

// Start starts the container if not already running
func (pc *PostgresContainer) Start(ctx context.Context) error {
	return pc.container.Start(ctx)
}
