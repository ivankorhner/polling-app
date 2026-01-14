package server

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/ivankorhner/polling-app/internal/config"
	"github.com/ivankorhner/polling-app/internal/server/middleware"
)

func AddRoutes(
	ctx context.Context,
	config *config.Config,
	logger *slog.Logger,
) http.Handler {
	mux := http.NewServeMux()

	middlewareDefaults := middleware.NewDefaults(ctx, config, logger)
}
