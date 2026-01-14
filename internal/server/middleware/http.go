package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/ivankorhner/polling-app/internal/logging"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

func httpRequest(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		ctx := r.Context()
		ctx = logging.AppendCtx(ctx,
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
		)
		r = r.WithContext(ctx)

		logger.LogAttrs(
			r.Context(),
			slog.LevelInfo,
			"started request",
		)

		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)

		logger.LogAttrs(
			r.Context(),
			slog.LevelInfo,
			"completed request",
			slog.Int("status", rw.statusCode),
			slog.Int("size", rw.size),
			slog.Duration("duration", time.Since(start)),
		)
	})
}
