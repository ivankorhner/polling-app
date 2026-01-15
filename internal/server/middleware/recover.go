package middleware

import (
	"log/slog"
	"net/http"
)

func panicRecovery(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.LogAttrs(
					r.Context(),
					slog.LevelError,
					"panic recovered",
					slog.String(RequestIDKey, requestIDFromContext(r.Context())),

					slog.String("path", r.URL.Path),
					slog.Any("panic", rec),
				)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
