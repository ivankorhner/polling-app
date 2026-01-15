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
		Save(ctx)
	require.NoError(t, err)

	_, err = testDB.Client.PollOption.Create().
		SetPollID(poll.ID).
		SetText("Option 2").
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

func TestHandleVote_Validation(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		pathID     string
		wantStatus int
		wantError  string
	}{
		{
			name:       "missing option_id",
			body:       `{"user_id": 1}`,
			pathID:     "1",
			wantStatus: http.StatusBadRequest,
			wantError:  "option_id is required",
		},
		{
			name:       "missing user_id",
			body:       `{"option_id": 1}`,
			pathID:     "1",
			wantStatus: http.StatusBadRequest,
			wantError:  "user_id is required",
		},
		{
			name:       "invalid poll ID",
			body:       `{"option_id": 1, "user_id": 1}`,
			pathID:     "invalid",
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid poll id",
		},
		{
			name:       "invalid JSON",
			body:       `{invalid}`,
			pathID:     "1",
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid request body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			testDB := testutil.SetupTestDB(ctx, t)
			defer testDB.Teardown(ctx)

			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

			req := httptest.NewRequest(http.MethodPost, "/polls/"+tt.pathID+"/vote", bytes.NewBufferString(tt.body))
			req.SetPathValue("id", tt.pathID)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler := server.HandleVote(logger, testDB.Client)
			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)

			var errResp server.ErrorResponse
			err := json.Unmarshal(rec.Body.Bytes(), &errResp)
			require.NoError(t, err)
			assert.Equal(t, tt.wantError, errResp.Error)
		})
	}
}

func TestHandleVote_NotFound(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(ctx context.Context, testDB *testutil.TestDB) (pollID int, optionID int, userID int)
		wantStatus int
		wantError  string
	}{
		{
			name: "poll not found",
			setup: func(ctx context.Context, testDB *testutil.TestDB) (int, int, int) {
				return 99999, 1, 1
			},
			wantStatus: http.StatusNotFound,
			wantError:  "poll not found",
		},
		{
			name: "option not found",
			setup: func(ctx context.Context, testDB *testutil.TestDB) (int, int, int) {
				user, _ := testDB.Client.User.Create().
					SetUsername("voter").
					SetEmail("voter@example.com").
					Save(ctx)
				poll, _ := testDB.Client.Poll.Create().
					SetOwnerID(user.ID).
					SetTitle("Test Poll").
					Save(ctx)
				return poll.ID, 99999, user.ID
			},
			wantStatus: http.StatusBadRequest,
			wantError:  "option not found or does not belong to poll",
		},
		{
			name: "user not found",
			setup: func(ctx context.Context, testDB *testutil.TestDB) (int, int, int) {
				owner, _ := testDB.Client.User.Create().
					SetUsername("owner").
					SetEmail("owner@example.com").
					Save(ctx)
				poll, _ := testDB.Client.Poll.Create().
					SetOwnerID(owner.ID).
					SetTitle("Test Poll").
					Save(ctx)
				option, _ := testDB.Client.PollOption.Create().
					SetPollID(poll.ID).
					SetText("Option 1").
					Save(ctx)
				return poll.ID, option.ID, 99999
			},
			wantStatus: http.StatusBadRequest,
			wantError:  "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			testDB := testutil.SetupTestDB(ctx, t)
			defer testDB.Teardown(ctx)

			pollID, optionID, userID := tt.setup(ctx, testDB)

			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

			body := fmt.Sprintf(`{"option_id": %d, "user_id": %d}`, optionID, userID)
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/polls/%d/vote", pollID), bytes.NewBufferString(body))
			req.SetPathValue("id", fmt.Sprintf("%d", pollID))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler := server.HandleVote(logger, testDB.Client)
			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)

			var errResp server.ErrorResponse
			err := json.Unmarshal(rec.Body.Bytes(), &errResp)
			require.NoError(t, err)
			assert.Equal(t, tt.wantError, errResp.Error)
		})
	}
}
