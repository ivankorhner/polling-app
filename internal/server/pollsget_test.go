//go:build integration

package server_test

import (
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

func TestHandleListPolls_EmptyDatabase(t *testing.T) {
	ctx := context.Background()
	testDB := testutil.SetupTestDB(ctx, t)
	defer testDB.Teardown(ctx)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	req := httptest.NewRequest(http.MethodGet, "/polls", nil)
	rec := httptest.NewRecorder()

	handler := server.HandleListPolls(logger, testDB.Client)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var polls []server.PollResponse
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
	opt1, err := testDB.Client.PollOption.Create().
		SetPollID(poll.ID).
		SetText("Option 1").
		Save(ctx)
	require.NoError(t, err)

	opt2, err := testDB.Client.PollOption.Create().
		SetPollID(poll.ID).
		SetText("Option 2").
		Save(ctx)
	require.NoError(t, err)

	// Create votes to test dynamic vote counting
	_, err = testDB.Client.Vote.Create().
		SetPollID(poll.ID).
		SetOptionID(opt1.ID).
		SetUserID(user.ID).
		Save(ctx)
	require.NoError(t, err)

	// Create another user and vote
	user2, err := testDB.Client.User.Create().
		SetUsername("testuser2").
		SetEmail("test2@example.com").
		Save(ctx)
	require.NoError(t, err)

	_, err = testDB.Client.Vote.Create().
		SetPollID(poll.ID).
		SetOptionID(opt2.ID).
		SetUserID(user2.ID).
		Save(ctx)
	require.NoError(t, err)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	req := httptest.NewRequest(http.MethodGet, "/polls", nil)
	rec := httptest.NewRecorder()

	handler := server.HandleListPolls(logger, testDB.Client)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var polls []server.PollResponse
	err = json.Unmarshal(rec.Body.Bytes(), &polls)
	require.NoError(t, err)
	require.Len(t, polls, 1)

	assert.Equal(t, poll.ID, polls[0].ID)
	assert.Equal(t, "Test Poll", polls[0].Title)
	require.Len(t, polls[0].Options, 2)

	// Check vote counts are calculated correctly
	for _, opt := range polls[0].Options {
		if opt.ID == opt1.ID {
			assert.Equal(t, 1, opt.VoteCount)
		} else if opt.ID == opt2.ID {
			assert.Equal(t, 1, opt.VoteCount)
		}
	}
}

func TestHandleGetPoll(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(ctx context.Context, testDB *testutil.TestDB) int
		pathID     string
		wantStatus int
		checkBody  func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name: "success",
			setup: func(ctx context.Context, testDB *testutil.TestDB) int {
				user, _ := testDB.Client.User.Create().
					SetUsername("testuser").
					SetEmail("test@example.com").
					Save(ctx)
				poll, _ := testDB.Client.Poll.Create().
					SetOwnerID(user.ID).
					SetTitle("Test Poll").
					Save(ctx)
				testDB.Client.PollOption.Create().
					SetPollID(poll.ID).
					SetText("Option 1").
					Save(ctx)
				return poll.ID
			},
			pathID:     "", // Will be set from setup
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var result server.PollResponse
				err := json.Unmarshal(rec.Body.Bytes(), &result)
				require.NoError(t, err)
				assert.Equal(t, "Test Poll", result.Title)
				require.Len(t, result.Options, 1)
				assert.Equal(t, "Option 1", result.Options[0].Text)
			},
		},
		{
			name: "not found",
			setup: func(ctx context.Context, testDB *testutil.TestDB) int {
				return 99999
			},
			wantStatus: http.StatusNotFound,
			checkBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var errResp server.ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &errResp)
				require.NoError(t, err)
				assert.Equal(t, "poll not found", errResp.Error)
			},
		},
		{
			name: "invalid ID",
			setup: func(ctx context.Context, testDB *testutil.TestDB) int {
				return 0 // Will use "invalid" as path
			},
			pathID:     "invalid",
			wantStatus: http.StatusBadRequest,
			checkBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var errResp server.ErrorResponse
				err := json.Unmarshal(rec.Body.Bytes(), &errResp)
				require.NoError(t, err)
				assert.Equal(t, "invalid poll id", errResp.Error)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			testDB := testutil.SetupTestDB(ctx, t)
			defer testDB.Teardown(ctx)

			pollID := tt.setup(ctx, testDB)
			pathID := tt.pathID
			if pathID == "" {
				pathID = fmt.Sprintf("%d", pollID)
			}

			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

			req := httptest.NewRequest(http.MethodGet, "/polls/"+pathID, nil)
			req.SetPathValue("id", pathID)
			rec := httptest.NewRecorder()

			handler := server.HandleGetPoll(logger, testDB.Client)
			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			tt.checkBody(t, rec)
		})
	}
}
