package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/ivankorhner/polling-app/internal/ent"
)

// RegisterUserRequest represents the request body for user registration
type RegisterUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

// UserResponse represents the response for user operations
type UserResponse struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// HandleRegisterUser handles user registration
func HandleRegisterUser(logger *slog.Logger, client *ent.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.LogAttrs(r.Context(), slog.LevelInfo, "register user: starting")

		// Limit request body size
		r.Body = http.MaxBytesReader(w, r.Body, MaxRequestBodySize)

		var req RegisterUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeValidationError(w, "invalid request body")
			return
		}

		// Validate username
		if errMsg := ValidateUsername(req.Username); errMsg != "" {
			writeValidationError(w, errMsg)
			return
		}

		// Validate email
		if errMsg := ValidateEmail(req.Email); errMsg != "" {
			writeValidationError(w, errMsg)
			return
		}

		// Normalize inputs
		req.Username = strings.TrimSpace(req.Username)
		req.Email = strings.TrimSpace(strings.ToLower(req.Email))

		user, err := client.User.Create().
			SetUsername(req.Username).
			SetEmail(req.Email).
			Save(r.Context())
		if err != nil {
			if ent.IsConstraintError(err) {
				writeConflictError(w, "username or email already exists")
				return
			}
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"failed to create user",
				slog.String("error", err.Error()),
			)
			writeInternalError(w, "failed to create user")
			return
		}

		logger.LogAttrs(
			r.Context(),
			slog.LevelInfo,
			"register user: completed",
			slog.Int("user_id", user.ID),
			slog.String("username", user.Username),
		)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(mapUserToResponse(user)); err != nil {
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"failed to encode user response",
				slog.String("error", err.Error()),
			)
		}
	})
}

func mapUserToResponse(u *ent.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
	}
}
