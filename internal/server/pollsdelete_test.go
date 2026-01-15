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

func TestHandleDeletePoll(t *testing.T) {
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
			wantStatus: http.StatusNoContent,
			checkBody:  nil,
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
				return 0
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

			req := httptest.NewRequest(http.MethodDelete, "/polls/"+pathID, nil)
			req.SetPathValue("id", pathID)
			rec := httptest.NewRecorder()

			handler := server.HandleDeletePoll(logger, testDB.Client)
			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.checkBody != nil {
				tt.checkBody(t, rec)
			}
		})
	}
}

func TestHandleDeletePoll_CascadeDelete(t *testing.T) {
	ctx := context.Background()
	testDB := testutil.SetupTestDB(ctx, t)
	defer testDB.Teardown(ctx)

	// Create user
	user, err := testDB.Client.User.Create().
		SetUsername("testuser").
		SetEmail("test@example.com").
		Save(ctx)
	require.NoError(t, err)

	// Create poll with options and votes
	poll, err := testDB.Client.Poll.Create().
		SetOwnerID(user.ID).
		SetTitle("Test Poll").
		Save(ctx)
	require.NoError(t, err)

	option, err := testDB.Client.PollOption.Create().
		SetPollID(poll.ID).
		SetText("Option 1").
		Save(ctx)
	require.NoError(t, err)

	_, err = testDB.Client.Vote.Create().
		SetPollID(poll.ID).
		SetOptionID(option.ID).
		SetUserID(user.ID).
		Save(ctx)
	require.NoError(t, err)

	// Verify entities exist
	pollCount, _ := testDB.Client.Poll.Query().Count(ctx)
	optionCount, _ := testDB.Client.PollOption.Query().Count(ctx)
	voteCount, _ := testDB.Client.Vote.Query().Count(ctx)
	assert.Equal(t, 1, pollCount)
	assert.Equal(t, 1, optionCount)
	assert.Equal(t, 1, voteCount)

	// Delete poll
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/polls/%d", poll.ID), nil)
	req.SetPathValue("id", fmt.Sprintf("%d", poll.ID))
	rec := httptest.NewRecorder()

	handler := server.HandleDeletePoll(logger, testDB.Client)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)

	// Verify all related entities are deleted
	pollCount, _ = testDB.Client.Poll.Query().Count(ctx)
	optionCount, _ = testDB.Client.PollOption.Query().Count(ctx)
	voteCount, _ = testDB.Client.Vote.Query().Count(ctx)
	assert.Equal(t, 0, pollCount)
	assert.Equal(t, 0, optionCount)
	assert.Equal(t, 0, voteCount)

	// User should still exist
	userCount, _ := testDB.Client.User.Query().Count(ctx)
	assert.Equal(t, 1, userCount)
}
