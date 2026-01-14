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
) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		// Correct order: panic recovery outermost, then request tracking, then timeout
		return panicRecovery(logger,
			requestId(logger,
				httpRequest(logger,
					timeout(config.ApiTimeout, h),
				),
			),
		)
	}
}
