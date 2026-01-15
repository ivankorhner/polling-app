package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/ivankorhner/polling-app/internal/ent"
	entpoll "github.com/ivankorhner/polling-app/internal/ent/poll"
	"github.com/ivankorhner/polling-app/internal/ent/polloption"
	"github.com/ivankorhner/polling-app/internal/ent/user"
)

type VoteRequest struct {
	OptionID int `json:"option_id"`
	UserID   int `json:"user_id"`
}

func HandleVote(logger *slog.Logger, client *ent.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pollIDStr := r.PathValue("id")
		pollID, err := strconv.Atoi(pollIDStr)
		if err != nil {
			http.Error(w, "invalid poll id", http.StatusBadRequest)
			return
		}

		var req VoteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.OptionID == 0 {
			http.Error(w, "option_id is required", http.StatusBadRequest)
			return
		}
		if req.UserID == 0 {
			http.Error(w, "user_id is required", http.StatusBadRequest)
			return
		}

		// Verify poll exists
		pollExists, err := client.Poll.Query().Where(entpoll.ID(pollID)).Exist(r.Context())
		if err != nil {
			logger.LogAttrs(r.Context(), slog.LevelError, "failed to check poll", slog.String("error", err.Error()))
			http.Error(w, "failed to submit vote", http.StatusInternalServerError)
			return
		}
		if !pollExists {
			http.Error(w, "poll not found", http.StatusNotFound)
			return
		}

		// Verify option exists and belongs to poll
		option, err := client.PollOption.Query().
			Where(polloption.ID(req.OptionID), polloption.PollID(pollID)).
			Only(r.Context())
		if err != nil {
			if ent.IsNotFound(err) {
				http.Error(w, "option not found or does not belong to poll", http.StatusBadRequest)
				return
			}
			logger.LogAttrs(r.Context(), slog.LevelError, "failed to check option", slog.String("error", err.Error()))
			http.Error(w, "failed to submit vote", http.StatusInternalServerError)
			return
		}

		// Verify user exists
		userExists, err := client.User.Query().Where(user.ID(req.UserID)).Exist(r.Context())
		if err != nil {
			logger.LogAttrs(r.Context(), slog.LevelError, "failed to check user", slog.String("error", err.Error()))
			http.Error(w, "failed to submit vote", http.StatusInternalServerError)
			return
		}
		if !userExists {
			http.Error(w, "user not found", http.StatusBadRequest)
			return
		}

		// Create vote and increment vote count in transaction
		err = createVoteWithIncrement(r.Context(), client, pollID, req.OptionID, req.UserID, option.VoteCount)
		if err != nil {
			if ent.IsConstraintError(err) {
				http.Error(w, "user has already voted on this poll", http.StatusConflict)
				return
			}
			logger.LogAttrs(r.Context(), slog.LevelError, "failed to create vote", slog.String("error", err.Error()))
			http.Error(w, "failed to submit vote", http.StatusInternalServerError)
			return
		}

		// Return updated poll
		poll, err := client.Poll.Query().
			Where(entpoll.ID(pollID)).
			WithOptions().
			Only(r.Context())
		if err != nil {
			logger.LogAttrs(r.Context(), slog.LevelError, "failed to reload poll", slog.String("error", err.Error()))
			http.Error(w, "failed to submit vote", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(mapPollToResponse(poll)); err != nil {
			logger.LogAttrs(r.Context(), slog.LevelError, "failed to encode response", slog.String("error", err.Error()))
		}
	})
}

func createVoteWithIncrement(ctx context.Context, client *ent.Client, pollID, optionID, userID, currentVoteCount int) error {
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
		tx.Rollback()
		return err
	}

	err = tx.PollOption.UpdateOneID(optionID).
		SetVoteCount(currentVoteCount + 1).
		Exec(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
