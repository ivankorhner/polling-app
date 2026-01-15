package server

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/ivankorhner/polling-app/internal/config"
	"github.com/ivankorhner/polling-app/internal/ent"
	"github.com/ivankorhner/polling-app/internal/handlers"
	"github.com/ivankorhner/polling-app/internal/server/middleware"
)

func AddRoutes(
	ctx context.Context,
	config *config.Config,
	logger *slog.Logger,
	client *ent.Client,
) http.Handler {
	mux := http.NewServeMux()

	middlewares := middleware.NewDefaults(ctx, config, logger)

	mux.Handle(http.MethodGet+" /health", handlers.HandleHealth(logger))
	mux.Handle(http.MethodGet+" /polls", handlers.HandleListPolls(logger, client))
	mux.Handle(http.MethodGet+" /polls/{id}", handlers.HandleGetPoll(logger, client))

	mux.Handle("/", http.NotFoundHandler())

	return middlewares(mux)
}
