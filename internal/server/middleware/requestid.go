package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/ivankorhner/polling-app/internal/logging"
)

const RequestIdKey string = "request_id"

func requestIdFromContext(ctx context.Context) string {
	if reqId, ok := ctx.Value(RequestIdKey).(string); ok {
		return reqId
	}
	return ""
}

func requestId(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if existing := requestIdFromContext(ctx); existing == "" {
			reqId := generateRequestId()
			ctx = logging.AppendCtx(ctx, slog.String(RequestIdKey, reqId))
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func generateRequestId() string {
	return uuid.New().String()
}
