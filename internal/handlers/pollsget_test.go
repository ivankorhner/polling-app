//go:build integration

package handlers_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ivankorhner/polling-app/internal/handlers"
	"github.com/ivankorhner/polling-app/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleListPolls_EmptyDatabase(t *testing.T) {
	ctx := context.Background()
	testDB := testutil.SetupTestDB(ctx, t)
	defer testDB.Teardown(ctx)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	req := httptest.NewRequest(http.MethodGet, "/polls", nil)
	rec := httptest.NewRecorder()

	handler := handlers.HandleListPolls(logger, testDB.Client)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var polls []handlers.PollResponse
	err := json.Unmarshal(rec.Body.Bytes(), &polls)
	require.NoError(t, err)
	assert.Empty(t, polls)
}

func TestHandleListPolls_WithPolls(t *testing.T) {
	ctx := context.Background()
	testDB := testutil.SetupTestDB(ctx, t)
	defer testDB.Teardown(ctx)

	// Create test user
	user, err := testDB.Client.User.Create().
		SetUsername("testuser").
		SetEmail("test@example.com").
		Save(ctx)
	require.NoError(t, err)

	// Create test poll
	poll, err := testDB.Client.Poll.Create().
		SetOwnerID(user.ID).
		SetTitle("Test Poll").
		Save(ctx)
	require.NoError(t, err)

	// Create test options
	_, err = testDB.Client.PollOption.Create().
		SetPollID(poll.ID).
		SetText("Option 1").
		SetVoteCount(5).
		Save(ctx)
	require.NoError(t, err)

	_, err = testDB.Client.PollOption.Create().
		SetPollID(poll.ID).
		SetText("Option 2").
		SetVoteCount(3).
		Save(ctx)
	require.NoError(t, err)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	req := httptest.NewRequest(http.MethodGet, "/polls", nil)
	rec := httptest.NewRecorder()

	handler := handlers.HandleListPolls(logger, testDB.Client)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var polls []handlers.PollResponse
	err = json.Unmarshal(rec.Body.Bytes(), &polls)
	require.NoError(t, err)
	require.Len(t, polls, 1)

	assert.Equal(t, poll.ID, polls[0].ID)
	assert.Equal(t, "Test Poll", polls[0].Title)
	require.Len(t, polls[0].Options, 2)
	assert.Equal(t, "Option 1", polls[0].Options[0].Text)
	assert.Equal(t, 5, polls[0].Options[0].VoteCount)
	assert.Equal(t, "Option 2", polls[0].Options[1].Text)
	assert.Equal(t, 3, polls[0].Options[1].VoteCount)
}
