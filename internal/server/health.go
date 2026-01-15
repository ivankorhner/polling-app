package server

import (
	"database/sql"
	"log/slog"
	"net/http"
)

// HandleHealth returns a health check handler that verifies database connectivity
func HandleHealth(logger *slog.Logger, db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := db.PingContext(r.Context()); err != nil {
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"health check failed: database unavailable",
				slog.String("error", err.Error()),
			)
			writeError(w, "database unavailable", ErrCodeInternal, http.StatusServiceUnavailable)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"status":"ok"}`))
		if err != nil {
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"failed to write health response",
				slog.String("error", err.Error()),
			)
		}
	})
}
