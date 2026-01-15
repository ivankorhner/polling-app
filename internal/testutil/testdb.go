package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/ivankorhner/polling-app/internal/ent"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	testDBUser     = "test"
	testDBPassword = "test"
	testDBName     = "test_db"
)

// TestDB holds the container and client for a test database
type TestDB struct {
	Container testcontainers.Container
	Client    *ent.Client
}

// SetupTestDB creates a new PostgreSQL container and returns an Ent client.
// The container is ephemeral with no volumes.
func SetupTestDB(ctx context.Context, t *testing.T) *TestDB {
	t.Helper()

	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase(testDBName),
		postgres.WithUsername(testDBUser),
		postgres.WithPassword(testDBPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("failed to get connection string: %v", err)
	}

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("failed to open database: %v", err)
	}

	drv := entsql.OpenDB(dialect.Postgres, db)
	client := ent.NewClient(ent.Driver(drv))

	// Run auto-migration to create schema
	if err := client.Schema.Create(ctx); err != nil {
		client.Close()
		container.Terminate(ctx)
		t.Fatalf("failed to run migrations: %v", err)
	}

	return &TestDB{
		Container: container,
		Client:    client,
	}
}

// Teardown cleans up the test database container
func (tdb *TestDB) Teardown(ctx context.Context) error {
	if tdb.Client != nil {
		tdb.Client.Close()
	}
	if tdb.Container != nil {
		return tdb.Container.Terminate(ctx)
	}
	return nil
}

// ConnectionString returns the database connection string
func (tdb *TestDB) ConnectionString(ctx context.Context) (string, error) {
	if tdb.Container == nil {
		return "", fmt.Errorf("container not initialized")
	}
	return tdb.Container.(interface {
		ConnectionString(context.Context, ...string) (string, error)
	}).ConnectionString(ctx, "sslmode=disable")
}
