package server

import (
	// "encoding/json"
	"log/slog"
	"net/http"
	// "strconv"
	// "time"
	// "github.com/go-playground/validator/v10"
	// "github.com/ivankorhner/polling-app/internal/polls"
)

func handleHealth(logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("OK"))
		if err != nil {
			logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"failed to write health response",
				slog.String("error", err.Error()),
			)
			// Note: Can't call http.Error here as headers are already written
			// Just log the error
			return
		}
	})
}

// type CreatePollRequest struct {
// 	Question string   `json:"question" validate:"required,min=5,max=500"`
// 	Options  []string `json:"options" validate:"required,min=2,max=10,dive,required,min=1,max=100"`
// }

// type OptionResponse struct {
// 	Id   int    `json:"id"`
// 	Text string `json:"text"`
// }

// type VoteResponse struct {
// 	OptionId  int       `json:"option_id"`
// 	UserEmail string    `json:"user_email"`
// 	Created   time.Time `json:"created_at"`
// }

// type PollResponse struct {
// 	ID       int64            `json:"id"`
// 	Question string           `json:"question"`
// 	Options  []OptionResponse `json:"options"`
// 	Votes    []VoteResponse   `json:"votes"`
// 	Created  time.Time        `json:"created_at"`
// }

// type VoteRequest struct {
// 	OptionId  int    `json:"option_id" validate:"required,min=0"`
// 	UserEmail string `json:"user_email" validate:"required,email"`
// }

// // CreatePoll handles POST /polls
// func (h *Handlers) CreatePoll(w http.ResponseWriter, r *http.Request) {
// 	var req CreatePollRequest
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		http.Error(w, "Invalid request body", http.StatusBadRequest)
// 		return
// 	}

// 	poll, err := h.service.CreatePoll(req)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusCreated)
// 	json.NewEncoder(w).Encode(poll)
// }

// // GetPoll handles GET /polls/{id}
// func (h *Handlers) GetPoll(w http.ResponseWriter, r *http.Request) {
// 	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
// 	if err != nil {
// 		http.Error(w, "Invalid poll ID", http.StatusBadRequest)
// 		return
// 	}

// 	poll, err := h.service.GetPoll(id)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusNotFound)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(poll)
// }

// // ListPolls handles GET /polls
// func (h *Handlers) ListPolls(w http.ResponseWriter, r *http.Request) {
// 	polls := h.service.ListPolls()

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(polls)
// }

// // DeletePoll handles DELETE /polls/{id}
// func (h *Handlers) DeletePoll(w http.ResponseWriter, r *http.Request) {
// 	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
// 	if err != nil {
// 		http.Error(w, "Invalid poll ID", http.StatusBadRequest)
// 		return
// 	}

// 	if err := h.service.DeletePoll(id); err != nil {
// 		http.Error(w, err.Error(), http.StatusNotFound)
// 		return
// 	}

// 	w.WriteHeader(http.StatusNoContent)
// }

// // Vote handles POST /polls/{id}/vote
// func (h *Handlers) Vote(w http.ResponseWriter, r *http.Request) {
// 	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
// 	if err != nil {
// 		http.Error(w, "Invalid poll ID", http.StatusBadRequest)
// 		return
// 	}

// 	var req VoteRequest
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		http.Error(w, "Invalid request body", http.StatusBadRequest)
// 		return
// 	}

// 	poll, err := h.service.Vote(id, req)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(poll)
// }
