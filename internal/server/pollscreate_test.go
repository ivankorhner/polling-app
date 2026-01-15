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

func TestHandleCreatePoll_Success(t *testing.T) {
	ctx := context.Background()
	testDB := testutil.SetupTestDB(ctx, t)
	defer testDB.Teardown(ctx)

	// Create owner user
	user, err := testDB.Client.User.Create().
		SetUsername("pollowner").
		SetEmail("owner@example.com").
		Save(ctx)
	require.NoError(t, err)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	body := fmt.Sprintf(`{"owner_id": %d, "title": "Favorite Color?", "options": ["Red", "Blue", "Green"]}`, user.ID)
	req := httptest.NewRequest(http.MethodPost, "/polls", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler := server.HandleCreatePoll(logger, testDB.Client)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var result server.PollResponse
	err = json.Unmarshal(rec.Body.Bytes(), &result)
	require.NoError(t, err)

	assert.Equal(t, "Favorite Color?", result.Title)
	assert.NotZero(t, result.ID)
	require.Len(t, result.Options, 3)
	assert.Equal(t, "Red", result.Options[0].Text)
	assert.Equal(t, "Blue", result.Options[1].Text)
	assert.Equal(t, "Green", result.Options[2].Text)
	// Verify vote counts start at 0
	for _, opt := range result.Options {
		assert.Equal(t, 0, opt.VoteCount)
	}
}

func TestHandleCreatePoll_Validation(t *testing.T) {
	tests := []struct {
		name       string
		setupUser  bool
		body       func(userID int) string
		wantStatus int
		wantError  string
	}{
		{
			name:      "missing title",
			setupUser: true,
			body: func(userID int) string {
				return fmt.Sprintf(`{"owner_id": %d, "options": ["A", "B"]}`, userID)
			},
			wantStatus: http.StatusBadRequest,
			wantError:  "title is required",
		},
		{
			name:      "missing owner_id",
			setupUser: false,
			body: func(_ int) string {
				return `{"title": "Test Poll", "options": ["A", "B"]}`
			},
			wantStatus: http.StatusBadRequest,
			wantError:  "owner_id is required",
		},
		{
			name:      "too few options",
			setupUser: true,
			body: func(userID int) string {
				return fmt.Sprintf(`{"owner_id": %d, "title": "Test Poll", "options": ["Only One"]}`, userID)
			},
			wantStatus: http.StatusBadRequest,
			wantError:  "at least 2 options are required",
		},
		{
			name:      "owner not found",
			setupUser: false,
			body: func(_ int) string {
				return `{"owner_id": 99999, "title": "Test Poll", "options": ["A", "B"]}`
			},
			wantStatus: http.StatusBadRequest,
			wantError:  "owner not found",
		},
		{
			name:      "invalid JSON",
			setupUser: false,
			body: func(_ int) string {
				return `{invalid json}`
			},
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid request body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			testDB := testutil.SetupTestDB(ctx, t)
			defer testDB.Teardown(ctx)

			var userID int
			if tt.setupUser {
				user, err := testDB.Client.User.Create().
					SetUsername("testuser").
					SetEmail("test@example.com").
					Save(ctx)
				require.NoError(t, err)
				userID = user.ID
			}

			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

			req := httptest.NewRequest(http.MethodPost, "/polls", bytes.NewBufferString(tt.body(userID)))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler := server.HandleCreatePoll(logger, testDB.Client)
			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)

			var errResp server.ErrorResponse
			err := json.Unmarshal(rec.Body.Bytes(), &errResp)
			require.NoError(t, err)
			assert.Equal(t, tt.wantError, errResp.Error)
		})
	}
}
