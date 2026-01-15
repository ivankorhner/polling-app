package server

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/ivankorhner/polling-app/internal/ent"
	entpoll "github.com/ivankorhner/polling-app/internal/ent/poll"
	"github.com/ivankorhner/polling-app/internal/ent/polloption"
	"github.com/ivankorhner/polling-app/internal/ent/user"
)

// VoteRequest represents the request body for voting
type VoteRequest struct {
	OptionID int `json:"option_id"`
	UserID   int `json:"user_id"`
}

// HandleVote handles vote submission
func HandleVote(logger *slog.Logger, client *ent.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Limit request body size
		r.Body = http.MaxBytesReader(w, r.Body, MaxRequestBodySize)

		pollIDStr := r.PathValue("id")
		pollID, err := strconv.Atoi(pollIDStr)
		if err != nil {
			writeValidationError(w, "invalid poll id")
			return
		}

		logger.LogAttrs(r.Context(), slog.LevelInfo, "submit vote: starting", slog.Int("poll_id", pollID))

		var req VoteRequest
		if decodeErr := json.NewDecoder(r.Body).Decode(&req); decodeErr != nil {
			writeValidationError(w, "invalid request body")
			return
		}

		if req.OptionID == 0 {
			writeValidationError(w, "option_id is required")
			return
		}
		if req.UserID == 0 {
			writeValidationError(w, "user_id is required")
			return
		}

		// Verify poll exists
		pollExists, err := client.Poll.Query().Where(entpoll.ID(pollID)).Exist(r.Context())
		if err != nil {
			logger.LogAttrs(r.Context(), slog.LevelError, "failed to check poll", slog.String("error", err.Error()))
			writeInternalError(w, "failed to submit vote")
			return
		}
		if !pollExists {
			writeNotFoundError(w, "poll not found")
			return
		}

		// Verify option exists and belongs to poll
		optionExists, err := client.PollOption.Query().
			Where(polloption.ID(req.OptionID), polloption.PollID(pollID)).
			Exist(r.Context())
		if err != nil {
			logger.LogAttrs(r.Context(), slog.LevelError, "failed to check option", slog.String("error", err.Error()))
			writeInternalError(w, "failed to submit vote")
			return
		}
		if !optionExists {
			writeValidationError(w, "option not found or does not belong to poll")
			return
		}

		// Verify user exists
		userExists, err := client.User.Query().Where(user.ID(req.UserID)).Exist(r.Context())
		if err != nil {
			logger.LogAttrs(r.Context(), slog.LevelError, "failed to check user", slog.String("error", err.Error()))
			writeInternalError(w, "failed to submit vote")
			return
		}
		if !userExists {
			writeValidationError(w, "user not found")
			return
		}

		// Create vote (no need to increment vote_count anymore - it's calculated dynamically)
		err = createVote(r.Context(), client, pollID, req.OptionID, req.UserID)
		if err != nil {
			if ent.IsConstraintError(err) {
				writeConflictError(w, "user has already voted on this poll")
				return
			}
			logger.LogAttrs(r.Context(), slog.LevelError, "failed to create vote", slog.String("error", err.Error()))
			writeInternalError(w, "failed to submit vote")
			return
		}

		// Return updated poll with vote counts
		poll, err := client.Poll.Query().
			Where(entpoll.ID(pollID)).
			WithOptions(func(q *ent.PollOptionQuery) {
				q.WithVotes()
			}).
			Only(r.Context())
		if err != nil {
			logger.LogAttrs(r.Context(), slog.LevelError, "failed to reload poll", slog.String("error", err.Error()))
			writeInternalError(w, "failed to submit vote")
			return
		}

		logger.LogAttrs(
			r.Context(),
			slog.LevelInfo,
			"submit vote: completed",
			slog.Int("poll_id", pollID),
			slog.Int("option_id", req.OptionID),
			slog.Int("user_id", req.UserID),
		)

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(mapPollToResponse(poll)); err != nil {
			logger.LogAttrs(r.Context(), slog.LevelError, "failed to encode response", slog.String("error", err.Error()))
		}
	})
}

func createVote(ctx context.Context, client *ent.Client, pollID, optionID, userID int) error {
	tx, err := client.Tx(ctx)
	if err != nil {
		return err
	}

	_, err = tx.Vote.Create().
		SetPollID(pollID).
		SetOptionID(optionID).
		SetUserID(userID).
		Save(ctx)
	if err != nil {
		return errors.Join(err, tx.Rollback())
	}

	return tx.Commit()
}
