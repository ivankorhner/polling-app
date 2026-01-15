package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"github.com/ivankorhner/polling-app/internal/logging"
)

// RequestIDKey is the context key for storing request IDs
const RequestIDKey string = "request_id"

func requestIDFromContext(ctx context.Context) string {
	if reqID, ok := ctx.Value(RequestIDKey).(string); ok {
		return reqID
	}
	return ""
}

func requestID(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if existing := requestIDFromContext(ctx); existing == "" {
			reqID := generateRequestID()
			ctx = logging.AppendCtx(ctx, slog.String(RequestIDKey, reqID))
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func generateRequestID() string {
	return uuid.New().String()
}
