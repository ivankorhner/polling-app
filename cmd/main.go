package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/ivankorhner/polling-app/internal/config"
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

	httpServer := &http.Server{
		Addr:     config.Addr(),
		Handler:  server.AddRoutes(ctx, config, logger),
		ErrorLog: slog.NewLogLogger(logger.Handler(), slog.LevelInfo),
	}

	serverErrors := make(chan error, 1)
	go func() {
		slog.LogAttrs(
			ctx,
			slog.LevelInfo,
			"server starting",
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
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	ctx := context.Background()
	config := config.LoadConfig()

	if err := run(ctx, config); err != nil {
		slog.LogAttrs(
			ctx,
			slog.LevelError,
			"application error",
			slog.Any("error", err),
		)
	}
}
