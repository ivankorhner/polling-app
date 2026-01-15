package server

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/ivankorhner/polling-app/internal/ent"
	entpoll "github.com/ivankorhner/polling-app/internal/ent/poll"
	"github.com/ivankorhner/polling-app/internal/ent/polloption"
	"github.com/ivankorhner/polling-app/internal/ent/vote"
)

func HandleDeletePoll(logger *slog.Logger, client *ent.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "invalid poll id", http.StatusBadRequest)
			return
		}

		// Check if poll exists
		exists, err := client.Poll.Query().Where(entpoll.ID(id)).Exist(r.Context())
		if err != nil {
			logger.LogAttrs(r.Context(), slog.LevelError, "failed to check poll", slog.String("error", err.Error()))
			http.Error(w, "failed to delete poll", http.StatusInternalServerError)
			return
		}
		if !exists {
			http.Error(w, "poll not found", http.StatusNotFound)
			return
		}

		// Delete poll and related entities in transaction
		if err := deletePollWithRelations(r.Context(), client, id); err != nil {
			logger.LogAttrs(r.Context(), slog.LevelError, "failed to delete poll", slog.String("error", err.Error()))
			http.Error(w, "failed to delete poll", http.StatusInternalServerError)
			return
		}

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
		_ = tx.Rollback()
		return err
	}

	// Delete poll options
	_, err = tx.PollOption.Delete().Where(polloption.PollID(pollID)).Exec(ctx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	// Delete the poll itself
	err = tx.Poll.DeleteOneID(pollID).Exec(ctx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
