package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/ivankorhner/polling-app/internal/ent"
	entpoll "github.com/ivankorhner/polling-app/internal/ent/poll"
	"github.com/ivankorhner/polling-app/internal/ent/user"
)

type CreatePollRequest struct {
	OwnerID int      `json:"owner_id"`
	Title   string   `json:"title"`
	Options []string `json:"options"`
}

func HandleCreatePoll(logger *slog.Logger, client *ent.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req CreatePollRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.Title == "" {
			http.Error(w, "title is required", http.StatusBadRequest)
			return
		}
		if req.OwnerID == 0 {
			http.Error(w, "owner_id is required", http.StatusBadRequest)
			return
		}
		if len(req.Options) < 2 {
			http.Error(w, "at least 2 options are required", http.StatusBadRequest)
			return
		}

		// Verify owner exists
		exists, err := client.User.Query().Where(user.ID(req.OwnerID)).Exist(r.Context())
		if err != nil {
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"failed to check user existence",
				slog.String("error", err.Error()),
			)
			http.Error(w, "failed to create poll", http.StatusInternalServerError)
			return
		}
		if !exists {
			http.Error(w, "owner not found", http.StatusBadRequest)
			return
		}

		poll, err := createPollWithOptions(r.Context(), client, req)
		if err != nil {
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"failed to create poll",
				slog.String("error", err.Error()),
			)
			http.Error(w, "failed to create poll", http.StatusInternalServerError)
			return
		}

		// Reload poll with options
		poll, err = client.Poll.Query().
			Where(entpoll.ID(poll.ID)).
			WithOptions().
			Only(r.Context())
		if err != nil {
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"failed to reload poll",
				slog.String("error", err.Error()),
			)
			http.Error(w, "failed to create poll", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(mapPollToResponse(poll)); err != nil {
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"failed to encode poll response",
				slog.String("error", err.Error()),
			)
		}
	})
}

func createPollWithOptions(ctx context.Context, client *ent.Client, req CreatePollRequest) (*ent.Poll, error) {
	tx, err := client.Tx(ctx)
	if err != nil {
		return nil, err
	}

	poll, err := tx.Poll.Create().
		SetOwnerID(req.OwnerID).
		SetTitle(req.Title).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	for _, optionText := range req.Options {
		_, err := tx.PollOption.Create().
			SetPollID(poll.ID).
			SetText(optionText).
			Save(ctx)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return poll, nil
}
