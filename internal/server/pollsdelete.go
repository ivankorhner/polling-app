package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/ivankorhner/polling-app/internal/ent"
	entpoll "github.com/ivankorhner/polling-app/internal/ent/poll"
	"github.com/ivankorhner/polling-app/internal/ent/polloption"
	"github.com/ivankorhner/polling-app/internal/ent/vote"
)

// HandleDeletePoll handles poll deletion
func HandleDeletePoll(logger *slog.Logger, client *ent.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			writeValidationError(w, "invalid poll id")
			return
		}

		logger.LogAttrs(r.Context(), slog.LevelInfo, "delete poll: starting", slog.Int("poll_id", id))

		// Check if poll exists
		exists, err := client.Poll.Query().Where(entpoll.ID(id)).Exist(r.Context())
		if err != nil {
			logger.LogAttrs(r.Context(), slog.LevelError, "failed to check poll", slog.String("error", err.Error()))
			writeInternalError(w, "failed to delete poll")
			return
		}
		if !exists {
			writeNotFoundError(w, "poll not found")
			return
		}

		// Delete poll and related entities in transaction
		if err := deletePollWithRelations(r.Context(), client, id); err != nil {
			logger.LogAttrs(r.Context(), slog.LevelError, "failed to delete poll", slog.String("error", err.Error()))
			writeInternalError(w, "failed to delete poll")
			return
		}

		logger.LogAttrs(r.Context(), slog.LevelInfo, "delete poll: completed", slog.Int("poll_id", id))

		w.WriteHeader(http.StatusNoContent)
	})
}

func deletePollWithRelations(ctx context.Context, client *ent.Client, pollID int) error {
	tx, err := client.Tx(ctx)
	if err != nil {
		return err
	}

	// Delete votes for this poll first
	_, err = tx.Vote.Delete().Where(vote.PollID(pollID)).Exec(ctx)
	if err != nil {
		return errors.Join(err, tx.Rollback())
	}

	// Delete poll options
	_, err = tx.PollOption.Delete().Where(polloption.PollID(pollID)).Exec(ctx)
	if err != nil {
		return errors.Join(err, tx.Rollback())
	}

	// Delete the poll itself
	err = tx.Poll.DeleteOneID(pollID).Exec(ctx)
	if err != nil {
		return errors.Join(err, tx.Rollback())
	}

	return tx.Commit()
}
