package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/ivankorhner/polling-app/internal/ent"
)

type RegisterUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

type UserResponse struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func HandleRegisterUser(logger *slog.Logger, client *ent.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req RegisterUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.Username == "" {
			http.Error(w, "username is required", http.StatusBadRequest)
			return
		}
		if req.Email == "" {
			http.Error(w, "email is required", http.StatusBadRequest)
			return
		}

		user, err := client.User.Create().
			SetUsername(req.Username).
			SetEmail(req.Email).
			Save(r.Context())
		if err != nil {
			if ent.IsConstraintError(err) {
				http.Error(w, "username or email already exists", http.StatusConflict)
				return
			}
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"failed to create user",
				slog.String("error", err.Error()),
			)
			http.Error(w, "failed to create user", http.StatusInternalServerError)
			return
		}

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
