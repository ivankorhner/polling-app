//go:build integration

package server_test

import (
	"context"
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

func TestHandleDeletePoll_Success(t *testing.T) {
	ctx := context.Background()
	testDB := testutil.SetupTestDB(ctx, t)
	defer testDB.Teardown(ctx)

	// Create owner
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
	_, err = testDB.Client.PollOption.Create().
		SetPollID(poll.ID).
		SetText("Option 1").
		SetVoteCount(0).
		Save(ctx)
	require.NoError(t, err)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/polls/%d", poll.ID), nil)
	req.SetPathValue("id", fmt.Sprintf("%d", poll.ID))
	rec := httptest.NewRecorder()

	handler := server.HandleDeletePoll(logger, testDB.Client)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)

	// Verify poll is deleted
	exists, err := testDB.Client.Poll.Query().Where().Exist(ctx)
	require.NoError(t, err)
	assert.False(t, exists)

	// Verify options are deleted
	optCount, err := testDB.Client.PollOption.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, optCount)

	// Verify owner still exists
	ownerExists, err := testDB.Client.User.Query().Exist(ctx)
	require.NoError(t, err)
	assert.True(t, ownerExists, "owner should not be deleted")
}

func TestHandleDeletePoll_WithVotes(t *testing.T) {
	ctx := context.Background()
	testDB := testutil.SetupTestDB(ctx, t)
	defer testDB.Teardown(ctx)

	// Create owner
	owner, err := testDB.Client.User.Create().
		SetUsername("owner").
		SetEmail("owner@example.com").
		Save(ctx)
	require.NoError(t, err)

	// Create voter
	voter, err := testDB.Client.User.Create().
		SetUsername("voter").
		SetEmail("voter@example.com").
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
		SetVoteCount(1).
		Save(ctx)
	require.NoError(t, err)

	// Create vote
	_, err = testDB.Client.Vote.Create().
		SetPollID(poll.ID).
		SetOptionID(option.ID).
		SetUserID(voter.ID).
		Save(ctx)
	require.NoError(t, err)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/polls/%d", poll.ID), nil)
	req.SetPathValue("id", fmt.Sprintf("%d", poll.ID))
	rec := httptest.NewRecorder()

	handler := server.HandleDeletePoll(logger, testDB.Client)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)

	// Verify poll is deleted
	pollExists, err := testDB.Client.Poll.Query().Exist(ctx)
	require.NoError(t, err)
	assert.False(t, pollExists)

	// Verify votes are deleted
	voteCount, err := testDB.Client.Vote.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, voteCount)

	// Verify options are deleted
	optCount, err := testDB.Client.PollOption.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, optCount)

	// Verify owner still exists
	ownerExists, err := testDB.Client.User.Query().Exist(ctx)
	require.NoError(t, err)
	assert.True(t, ownerExists, "users should not be deleted")

	// Verify both users still exist
	userCount, err := testDB.Client.User.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, userCount, "both owner and voter should still exist")
}

func TestHandleDeletePoll_NotFound(t *testing.T) {
	ctx := context.Background()
	testDB := testutil.SetupTestDB(ctx, t)
	defer testDB.Teardown(ctx)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	req := httptest.NewRequest(http.MethodDelete, "/polls/99999", nil)
	req.SetPathValue("id", "99999")
	rec := httptest.NewRecorder()

	handler := server.HandleDeletePoll(logger, testDB.Client)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandleDeletePoll_InvalidID(t *testing.T) {
	ctx := context.Background()
	testDB := testutil.SetupTestDB(ctx, t)
	defer testDB.Teardown(ctx)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	req := httptest.NewRequest(http.MethodDelete, "/polls/invalid", nil)
	req.SetPathValue("id", "invalid")
	rec := httptest.NewRecorder()

	handler := server.HandleDeletePoll(logger, testDB.Client)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
