package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/ivankorhner/polling-app/internal/config"
	"github.com/ivankorhner/polling-app/internal/ent"
	"github.com/ivankorhner/polling-app/internal/logging"
	"github.com/ivankorhner/polling-app/internal/server"
)

func run(
	ctx context.Context,
	config *config.Config,
) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	logger := logging.NewLogger(slog.LevelInfo)
	slog.SetDefault(logger)

	// Open database connection with pooling configuration
	db, err := sql.Open("pgx", config.DatabaseURL())
	if err != nil {
		return err
	}
	defer db.Close()

	// Configure connection pool
	db.SetMaxOpenConns(25)                  // Maximum number of open connections
	db.SetMaxIdleConns(5)                   // Maximum number of idle connections
	db.SetConnMaxLifetime(5 * time.Minute)  // Maximum lifetime of a connection
	db.SetConnMaxIdleTime(10 * time.Minute) // Maximum idle time of a connection

	// Create Ent driver with the configured connection pool
	drv := entsql.OpenDB(dialect.Postgres, db)
	client := ent.NewClient(ent.Driver(drv))
	defer client.Close()

	slog.LogAttrs(
		ctx,
		slog.LevelInfo,
		"database connection established",
		slog.String("host", config.DBHost),
		slog.Int("port", config.DBPort),
		slog.String("database", config.DBName),
	)

	httpServer := &http.Server{
		Addr:         config.Addr(),
		Handler:      server.AddRoutes(ctx, config, logger, db, client),
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelInfo),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		slog.LogAttrs(
			ctx,
			slog.LevelInfo,
			"server starting",
			slog.String("addr", config.Addr()),
		)
		serverErrors <- httpServer.ListenAndServe()
	}()

	// wait for interrupt or server error
	select {
	case err := <-serverErrors:
		return err
	case <-ctx.Done():
		slog.LogAttrs(
			ctx,
			slog.LevelInfo,
			"shutting down server",
		)
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	ctx := context.Background()
	cfg := config.LoadConfig()

	if err := run(ctx, cfg); err != nil {
		slog.LogAttrs(
			ctx,
			slog.LevelError,
			"application error",
			slog.Any("error", err),
		)
		os.Exit(1)
	}
}
