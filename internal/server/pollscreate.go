package server

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/ivankorhner/polling-app/internal/ent"
	entpoll "github.com/ivankorhner/polling-app/internal/ent/poll"
	"github.com/ivankorhner/polling-app/internal/ent/user"
)

// CreatePollRequest represents the request body for poll creation
type CreatePollRequest struct {
	OwnerID int      `json:"owner_id"`
	Title   string   `json:"title"`
	Options []string `json:"options"`
}

// HandleCreatePoll handles poll creation
func HandleCreatePoll(logger *slog.Logger, client *ent.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.LogAttrs(r.Context(), slog.LevelInfo, "create poll: starting")

		// Limit request body size
		r.Body = http.MaxBytesReader(w, r.Body, MaxRequestBodySize)

		var req CreatePollRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeValidationError(w, "invalid request body")
			return
		}

		// Validate title
		if errMsg := ValidatePollTitle(req.Title); errMsg != "" {
			writeValidationError(w, errMsg)
			return
		}

		// Validate owner_id
		if req.OwnerID == 0 {
			writeValidationError(w, "owner_id is required")
			return
		}

		// Validate options
		if errMsg := ValidatePollOptions(req.Options); errMsg != "" {
			writeValidationError(w, errMsg)
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
			writeInternalError(w, "failed to create poll")
			return
		}
		if !exists {
			writeValidationError(w, "owner not found")
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
			writeInternalError(w, "failed to create poll")
			return
		}

		// Reload poll with options and vote counts
		poll, err = client.Poll.Query().
			Where(entpoll.ID(poll.ID)).
			WithOptions(func(q *ent.PollOptionQuery) {
				q.WithVotes()
			}).
			Only(r.Context())
		if err != nil {
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"failed to reload poll",
				slog.String("error", err.Error()),
			)
			writeInternalError(w, "failed to create poll")
			return
		}

		logger.LogAttrs(
			r.Context(),
			slog.LevelInfo,
			"create poll: completed",
			slog.Int("poll_id", poll.ID),
			slog.String("title", poll.Title),
			slog.Int("options_count", len(poll.Edges.Options)),
		)

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
		return nil, errors.Join(err, tx.Rollback())
	}

	for _, optionText := range req.Options {
		_, err := tx.PollOption.Create().
			SetPollID(poll.ID).
			SetText(optionText).
			Save(ctx)
		if err != nil {
			return nil, errors.Join(err, tx.Rollback())
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return poll, nil
}
