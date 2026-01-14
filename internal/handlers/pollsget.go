package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/ivankorhner/polling-app/internal/ent"
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
