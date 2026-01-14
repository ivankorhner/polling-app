package handlers

import (
	"log/slog"
	"net/http"
)

func HandleHealth(logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("OK"))
		if err != nil {
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"failed to write health response",
				slog.String("error", err.Error()),
			)
			return
		}
	})
}
