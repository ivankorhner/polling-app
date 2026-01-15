package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/ivankorhner/polling-app/internal/ent"
	entpoll "github.com/ivankorhner/polling-app/internal/ent/poll"
)

// PollResponse represents the response for poll operations
type PollResponse struct {
	ID        int              `json:"id"`
	Title     string           `json:"title"`
	CreatedAt time.Time        `json:"created_at"`
	Options   []OptionResponse `json:"options"`
}

// OptionResponse represents a poll option in responses
type OptionResponse struct {
	ID        int    `json:"id"`
	Text      string `json:"text"`
	VoteCount int    `json:"vote_count"`
}

// HandleListPolls handles listing all polls
func HandleListPolls(logger *slog.Logger, client *ent.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.LogAttrs(r.Context(), slog.LevelInfo, "list polls: starting")

		polls, err := client.Poll.Query().
			WithOptions(func(q *ent.PollOptionQuery) {
				q.WithVotes()
			}).
			All(r.Context())
		if err != nil {
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"failed to query polls",
				slog.String("error", err.Error()),
			)
			writeInternalError(w, "failed to retrieve polls")
			return
		}

		response := make([]PollResponse, len(polls))
		for i, p := range polls {
			response[i] = mapPollToResponse(p)
		}

		logger.LogAttrs(
			r.Context(),
			slog.LevelInfo,
			"list polls: completed",
			slog.Int("count", len(polls)),
		)

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"failed to encode polls response",
				slog.String("error", err.Error()),
			)
		}
	})
}

// HandleGetPoll handles getting a single poll by ID
func HandleGetPoll(logger *slog.Logger, client *ent.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			writeValidationError(w, "invalid poll id")
			return
		}

		logger.LogAttrs(r.Context(), slog.LevelInfo, "get poll: starting", slog.Int("poll_id", id))

		poll, err := client.Poll.Query().
			WithOptions(func(q *ent.PollOptionQuery) {
				q.WithVotes()
			}).
			Where(entpoll.ID(id)).
			Only(r.Context())
		if err != nil {
			if ent.IsNotFound(err) {
				writeNotFoundError(w, "poll not found")
				return
			}
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"failed to query poll",
				slog.String("error", err.Error()),
				slog.Int("poll_id", id),
			)
			writeInternalError(w, "failed to retrieve poll")
			return
		}

		logger.LogAttrs(
			r.Context(),
			slog.LevelInfo,
			"get poll: completed",
			slog.Int("poll_id", poll.ID),
			slog.String("title", poll.Title),
		)

		w.Header().Set("Content-Type", "application/json")
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

func mapPollToResponse(p *ent.Poll) PollResponse {
	return PollResponse{
		ID:        p.ID,
		Title:     p.Title,
		CreatedAt: p.CreatedAt,
		Options:   mapOptionsToResponse(p.Edges.Options),
	}
}

func mapOptionsToResponse(options []*ent.PollOption) []OptionResponse {
	result := make([]OptionResponse, len(options))
	for i, o := range options {
		// Calculate vote count dynamically from the votes edge
		voteCount := len(o.Edges.Votes)
		result[i] = OptionResponse{
			ID:        o.ID,
			Text:      o.Text,
			VoteCount: voteCount,
		}
	}
	return result
}
