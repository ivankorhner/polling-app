package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/ivankorhner/polling-app/internal/config"
)

func NewDefaults(
	ctx context.Context,
	config *config.Config,
	logger *slog.Logger,
) http.Handler {
	return requestId(logger,
		httpRequest(logger,
			timeout(config.ApiTimeout,
				panicRecovery(logger, nil),
			),
		),
	)
}
