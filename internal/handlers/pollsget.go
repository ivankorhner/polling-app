package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/ivankorhner/polling-app/internal/ent"
	entpoll "github.com/ivankorhner/polling-app/internal/ent/poll"
)

type PollResponse struct {
	ID        int              `json:"id"`
	Title     string           `json:"title"`
	CreatedAt time.Time        `json:"created_at"`
	Options   []OptionResponse `json:"options"`
}

type OptionResponse struct {
	ID        int    `json:"id"`
	Text      string `json:"text"`
	VoteCount int    `json:"vote_count"`
}

func HandleListPolls(logger *slog.Logger, client *ent.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		polls, err := client.Poll.Query().
			WithOptions().
			All(r.Context())
		if err != nil {
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"failed to query polls",
				slog.String("error", err.Error()),
			)
			http.Error(w, "failed to retrieve polls", http.StatusInternalServerError)
			return
		}

		response := make([]PollResponse, len(polls))
		for i, p := range polls {
			response[i] = mapPollToResponse(p)
		}

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

func HandleGetPoll(logger *slog.Logger, client *ent.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "invalid poll id", http.StatusBadRequest)
			return
		}

		poll, err := client.Poll.Query().
			WithOptions().
			Where(entpoll.ID(id)).
			Only(r.Context())
		if err != nil {
			if ent.IsNotFound(err) {
				http.Error(w, "poll not found", http.StatusNotFound)
				return
			}
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"failed to query poll",
				slog.String("error", err.Error()),
				slog.Int("poll_id", id),
			)
			http.Error(w, "failed to retrieve poll", http.StatusInternalServerError)
			return
		}

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
		result[i] = OptionResponse{
			ID:        o.ID,
			Text:      o.Text,
			VoteCount: o.VoteCount,
		}
	}
	return result
}
