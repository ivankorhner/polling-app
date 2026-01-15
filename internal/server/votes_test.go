//go:build integration

package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ivankorhner/polling-app/internal/server"
	"github.com/ivankorhner/polling-app/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleVote_Success(t *testing.T) {
	ctx := context.Background()
	testDB := testutil.SetupTestDB(ctx, t)
	defer testDB.Teardown(ctx)

	// Create user
	user, err := testDB.Client.User.Create().
		SetUsername("voter").
		SetEmail("voter@example.com").
		Save(ctx)
	require.NoError(t, err)

	// Create poll owner
	owner, err := testDB.Client.User.Create().
		SetUsername("owner").
		SetEmail("owner@example.com").
		Save(ctx)
	require.NoError(t, err)

	// Create poll
	poll, err := testDB.Client.Poll.Create().
		SetOwnerID(owner.ID).
		SetTitle("Test Poll").
		Save(ctx)
	require.NoError(t, err)

	// Create options
	option1, err := testDB.Client.PollOption.Create().
		SetPollID(poll.ID).
		SetText("Option 1").
		SetVoteCount(0).
		Save(ctx)
	require.NoError(t, err)

	_, err = testDB.Client.PollOption.Create().
		SetPollID(poll.ID).
		SetText("Option 2").
		SetVoteCount(0).
		Save(ctx)
	require.NoError(t, err)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	body := fmt.Sprintf(`{"option_id": %d, "user_id": %d}`, option1.ID, user.ID)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/polls/%d/vote", poll.ID), bytes.NewBufferString(body))
	req.SetPathValue("id", fmt.Sprintf("%d", poll.ID))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler := server.HandleVote(logger, testDB.Client)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var result server.PollResponse
	err = json.Unmarshal(rec.Body.Bytes(), &result)
	require.NoError(t, err)

	assert.Equal(t, poll.ID, result.ID)
	require.Len(t, result.Options, 2)

	// Find the voted option by ID and verify vote count incremented
	var votedOption *server.OptionResponse
	for i := range result.Options {
		if result.Options[i].ID == option1.ID {
			votedOption = &result.Options[i]
			break
		}
	}
	require.NotNil(t, votedOption, "voted option should be in response")
	assert.Equal(t, 1, votedOption.VoteCount)
}

func TestHandleVote_DuplicateVote(t *testing.T) {
	ctx := context.Background()
	testDB := testutil.SetupTestDB(ctx, t)
	defer testDB.Teardown(ctx)

	// Create user
	user, err := testDB.Client.User.Create().
		SetUsername("voter").
		SetEmail("voter@example.com").
		Save(ctx)
	require.NoError(t, err)

	// Create poll
	poll, err := testDB.Client.Poll.Create().
		SetOwnerID(user.ID).
		SetTitle("Test Poll").
		Save(ctx)
	require.NoError(t, err)

	// Create option
	option, err := testDB.Client.PollOption.Create().
		SetPollID(poll.ID).
		SetText("Option 1").
		SetVoteCount(0).
		Save(ctx)
	require.NoError(t, err)

	// Create existing vote
	_, err = testDB.Client.Vote.Create().
		SetPollID(poll.ID).
		SetOptionID(option.ID).
		SetUserID(user.ID).
		Save(ctx)
	require.NoError(t, err)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	body := fmt.Sprintf(`{"option_id": %d, "user_id": %d}`, option.ID, user.ID)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/polls/%d/vote", poll.ID), bytes.NewBufferString(body))
	req.SetPathValue("id", fmt.Sprintf("%d", poll.ID))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler := server.HandleVote(logger, testDB.Client)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestHandleVote_PollNotFound(t *testing.T) {
	ctx := context.Background()
	testDB := testutil.SetupTestDB(ctx, t)
	defer testDB.Teardown(ctx)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	body := `{"option_id": 1, "user_id": 1}`
	req := httptest.NewRequest(http.MethodPost, "/polls/99999/vote", bytes.NewBufferString(body))
	req.SetPathValue("id", "99999")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler := server.HandleVote(logger, testDB.Client)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandleVote_OptionNotFound(t *testing.T) {
	ctx := context.Background()
	testDB := testutil.SetupTestDB(ctx, t)
	defer testDB.Teardown(ctx)

	// Create user
	user, err := testDB.Client.User.Create().
		SetUsername("voter").
		SetEmail("voter@example.com").
		Save(ctx)
	require.NoError(t, err)

	// Create poll
	poll, err := testDB.Client.Poll.Create().
		SetOwnerID(user.ID).
		SetTitle("Test Poll").
		Save(ctx)
	require.NoError(t, err)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	body := fmt.Sprintf(`{"option_id": 99999, "user_id": %d}`, user.ID)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/polls/%d/vote", poll.ID), bytes.NewBufferString(body))
	req.SetPathValue("id", fmt.Sprintf("%d", poll.ID))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler := server.HandleVote(logger, testDB.Client)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandleVote_UserNotFound(t *testing.T) {
	ctx := context.Background()
	testDB := testutil.SetupTestDB(ctx, t)
	defer testDB.Teardown(ctx)

	// Create poll owner
	owner, err := testDB.Client.User.Create().
		SetUsername("owner").
		SetEmail("owner@example.com").
		Save(ctx)
	require.NoError(t, err)

	// Create poll
	poll, err := testDB.Client.Poll.Create().
		SetOwnerID(owner.ID).
		SetTitle("Test Poll").
		Save(ctx)
	require.NoError(t, err)

	// Create option
	option, err := testDB.Client.PollOption.Create().
		SetPollID(poll.ID).
		SetText("Option 1").
		SetVoteCount(0).
		Save(ctx)
	require.NoError(t, err)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	body := fmt.Sprintf(`{"option_id": %d, "user_id": 99999}`, option.ID)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/polls/%d/vote", poll.ID), bytes.NewBufferString(body))
	req.SetPathValue("id", fmt.Sprintf("%d", poll.ID))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler := server.HandleVote(logger, testDB.Client)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandleVote_MissingOptionID(t *testing.T) {
	ctx := context.Background()
	testDB := testutil.SetupTestDB(ctx, t)
	defer testDB.Teardown(ctx)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	body := `{"user_id": 1}`
	req := httptest.NewRequest(http.MethodPost, "/polls/1/vote", bytes.NewBufferString(body))
	req.SetPathValue("id", "1")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler := server.HandleVote(logger, testDB.Client)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandleVote_MissingUserID(t *testing.T) {
	ctx := context.Background()
	testDB := testutil.SetupTestDB(ctx, t)
	defer testDB.Teardown(ctx)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	body := `{"option_id": 1}`
	req := httptest.NewRequest(http.MethodPost, "/polls/1/vote", bytes.NewBufferString(body))
	req.SetPathValue("id", "1")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler := server.HandleVote(logger, testDB.Client)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandleVote_InvalidPollID(t *testing.T) {
	ctx := context.Background()
	testDB := testutil.SetupTestDB(ctx, t)
	defer testDB.Teardown(ctx)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	body := `{"option_id": 1, "user_id": 1}`
	req := httptest.NewRequest(http.MethodPost, "/polls/invalid/vote", bytes.NewBufferString(body))
	req.SetPathValue("id", "invalid")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler := server.HandleVote(logger, testDB.Client)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
