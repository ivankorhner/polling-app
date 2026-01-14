package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleCreatePoll_Success(t *testing.T) {
	// Reset store for clean test
	store = &PollStore{
		polls:  make(map[int64]*Poll),
		nextID: 1,
	}

	req := CreatePollRequest{
		Question: "What is your favorite color?",
		Options:  []string{"Red", "Blue", "Green"},
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/polls", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handleCreatePoll(w, httpReq)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var resp Poll
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if resp.Question != req.Question {
		t.Errorf("Expected question %q, got %q", req.Question, resp.Question)
	}

	if len(resp.Options) != len(req.Options) {
		t.Errorf("Expected %d options, got %d", len(req.Options), len(resp.Options))
	}
}

func TestHandleCreatePoll_ValidationError(t *testing.T) {
	tests := []struct {
		name    string
		request CreatePollRequest
	}{
		{
			name: "Empty question",
			request: CreatePollRequest{
				Question: "",
				Options:  []string{"A", "B"},
			},
		},
		{
			name: "Question too short",
			request: CreatePollRequest{
				Question: "Hi?",
				Options:  []string{"A", "B"},
			},
		},
		{
			name: "Insufficient options",
			request: CreatePollRequest{
				Question: "Valid question here",
				Options:  []string{"A"},
			},
		},
		{
			name: "Empty option",
			request: CreatePollRequest{
				Question: "Valid question here",
				Options:  []string{"A", ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			httpReq := httptest.NewRequest("POST", "/polls", bytes.NewReader(body))
			httpReq.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			handleCreatePoll(w, httpReq)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
			}
		})
	}
}

func TestHandleCreatePoll_InvalidJSON(t *testing.T) {
	httpReq := httptest.NewRequest("POST", "/polls", bytes.NewReader([]byte("invalid json")))
	httpReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handleCreatePoll(w, httpReq)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandleListPolls(t *testing.T) {
	// Reset store and add a poll
	store = &PollStore{
		polls:  make(map[int64]*Poll),
		nextID: 1,
	}

	store.CreatePoll(CreatePollRequest{
		Question: "Test question?",
		Options:  []string{"Yes", "No"},
	})

	httpReq := httptest.NewRequest("GET", "/polls", nil)
	w := httptest.NewRecorder()
	handleListPolls(w, httpReq)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp []*Poll
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if len(resp) != 1 {
		t.Errorf("Expected 1 poll, got %d", len(resp))
	}
}

func TestHandleGetPoll_Success(t *testing.T) {
	// Reset store and add a poll
	store = &PollStore{
		polls:  make(map[int64]*Poll),
		nextID: 1,
	}

	poll := store.CreatePoll(CreatePollRequest{
		Question: "What's your favorite language?",
		Options:  []string{"Go", "Rust", "Python"},
	})

	httpReq := httptest.NewRequest("GET", "/polls/1", nil)
	httpReq.SetPathValue("id", "1")

	w := httptest.NewRecorder()
	handleGetPoll(w, httpReq)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp Poll
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if resp.ID != poll.ID {
		t.Errorf("Expected poll ID %d, got %d", poll.ID, resp.ID)
	}
}

func TestHandleGetPoll_NotFound(t *testing.T) {
	store = &PollStore{
		polls:  make(map[int64]*Poll),
		nextID: 1,
	}

	httpReq := httptest.NewRequest("GET", "/polls/999", nil)
	httpReq.SetPathValue("id", "999")

	w := httptest.NewRecorder()
	handleGetPoll(w, httpReq)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestHandleGetPoll_InvalidID(t *testing.T) {
	httpReq := httptest.NewRequest("GET", "/polls/invalid", nil)
	httpReq.SetPathValue("id", "invalid")

	w := httptest.NewRecorder()
	handleGetPoll(w, httpReq)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandleDeletePoll_Success(t *testing.T) {
	// Reset store and add a poll
	store = &PollStore{
		polls:  make(map[int64]*Poll),
		nextID: 1,
	}

	store.CreatePoll(CreatePollRequest{
		Question: "Test question?",
		Options:  []string{"Yes", "No"},
	})

	httpReq := httptest.NewRequest("DELETE", "/polls/1", nil)
	httpReq.SetPathValue("id", "1")

	w := httptest.NewRecorder()
	handleDeletePoll(w, httpReq)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
	}

	// Verify poll is deleted
	_, exists := store.GetPoll(1)
	if exists {
		t.Error("Poll should be deleted")
	}
}

func TestHandleDeletePoll_NotFound(t *testing.T) {
	store = &PollStore{
		polls:  make(map[int64]*Poll),
		nextID: 1,
	}

	httpReq := httptest.NewRequest("DELETE", "/polls/999", nil)
	httpReq.SetPathValue("id", "999")

	w := httptest.NewRecorder()
	handleDeletePoll(w, httpReq)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestHandleVote_Success(t *testing.T) {
	// Reset store and add a poll
	store = &PollStore{
		polls:  make(map[int64]*Poll),
		nextID: 1,
	}

	store.CreatePoll(CreatePollRequest{
		Question: "Favorite color?",
		Options:  []string{"Red", "Blue", "Green"},
	})

	voteReq := VoteRequest{
		OptionId:  0,
		UserEmail: "user@example.com",
	}

	body, _ := json.Marshal(voteReq)
	httpReq := httptest.NewRequest("POST", "/polls/1/vote", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.SetPathValue("id", "1")

	w := httptest.NewRecorder()
	handleVote(w, httpReq)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp Poll
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if resp.Votes[0] != 1 {
		t.Errorf("Expected 1 vote for option 0, got %d", resp.Votes[0])
	}
}

func TestHandleVote_ValidationError(t *testing.T) {
	// Reset store and add a poll
	store = &PollStore{
		polls:  make(map[int64]*Poll),
		nextID: 1,
	}

	store.CreatePoll(CreatePollRequest{
		Question: "Favorite color?",
		Options:  []string{"Red", "Blue"},
	})

	tests := []struct {
		name     string
		voteReq  VoteRequest
		pollID   string
	}{
		{
			name: "Invalid email",
			voteReq: VoteRequest{
				OptionId:  0,
				UserEmail: "not-an-email",
			},
			pollID: "1",
		},
		{
			name: "Missing email",
			voteReq: VoteRequest{
				OptionId:  0,
				UserEmail: "",
			},
			pollID: "1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.voteReq)
			httpReq := httptest.NewRequest("POST", "/polls/1/vote", bytes.NewReader(body))
			httpReq.Header.Set("Content-Type", "application/json")
			httpReq.SetPathValue("id", tt.pollID)

			w := httptest.NewRecorder()
			handleVote(w, httpReq)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
			}
		})
	}
}

func TestHandleVote_InvalidPollID(t *testing.T) {
	voteReq := VoteRequest{
		OptionId:  0,
		UserEmail: "user@example.com",
	}

	body, _ := json.Marshal(voteReq)
	httpReq := httptest.NewRequest("POST", "/polls/invalid/vote", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.SetPathValue("id", "invalid")

	w := httptest.NewRecorder()
	handleVote(w, httpReq)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandleVote_PollNotFound(t *testing.T) {
	store = &PollStore{
		polls:  make(map[int64]*Poll),
		nextID: 1,
	}

	voteReq := VoteRequest{
		OptionId:  0,
		UserEmail: "user@example.com",
	}

	body, _ := json.Marshal(voteReq)
	httpReq := httptest.NewRequest("POST", "/polls/999/vote", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.SetPathValue("id", "999")

	w := httptest.NewRecorder()
	handleVote(w, httpReq)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}
