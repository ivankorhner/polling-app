package middleware

import (
	"net/http"
	"time"
)

func timeout(timeout time.Duration, next http.Handler) http.Handler {
	return http.TimeoutHandler(next, timeout, "request took too long")
}
