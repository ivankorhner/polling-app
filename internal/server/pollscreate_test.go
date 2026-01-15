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
}

func TestHandleCreatePoll_MissingTitle(t *testing.T) {
	ctx := context.Background()
	testDB := testutil.SetupTestDB(ctx, t)
	defer testDB.Teardown(ctx)

	user, err := testDB.Client.User.Create().
		SetUsername("testuser").
		SetEmail("test@example.com").
		Save(ctx)
	require.NoError(t, err)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	body := fmt.Sprintf(`{"owner_id": %d, "options": ["A", "B"]}`, user.ID)
	req := httptest.NewRequest(http.MethodPost, "/polls", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler := server.HandleCreatePoll(logger, testDB.Client)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandleCreatePoll_MissingOwnerID(t *testing.T) {
	ctx := context.Background()
	testDB := testutil.SetupTestDB(ctx, t)
	defer testDB.Teardown(ctx)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	body := `{"title": "Test Poll", "options": ["A", "B"]}`
	req := httptest.NewRequest(http.MethodPost, "/polls", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler := server.HandleCreatePoll(logger, testDB.Client)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandleCreatePoll_TooFewOptions(t *testing.T) {
	ctx := context.Background()
	testDB := testutil.SetupTestDB(ctx, t)
	defer testDB.Teardown(ctx)

	user, err := testDB.Client.User.Create().
		SetUsername("testuser").
		SetEmail("test@example.com").
		Save(ctx)
	require.NoError(t, err)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	body := fmt.Sprintf(`{"owner_id": %d, "title": "Test Poll", "options": ["Only One"]}`, user.ID)
	req := httptest.NewRequest(http.MethodPost, "/polls", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler := server.HandleCreatePoll(logger, testDB.Client)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandleCreatePoll_OwnerNotFound(t *testing.T) {
	ctx := context.Background()
	testDB := testutil.SetupTestDB(ctx, t)
	defer testDB.Teardown(ctx)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	body := `{"owner_id": 99999, "title": "Test Poll", "options": ["A", "B"]}`
	req := httptest.NewRequest(http.MethodPost, "/polls", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler := server.HandleCreatePoll(logger, testDB.Client)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandleCreatePoll_InvalidJSON(t *testing.T) {
	ctx := context.Background()
	testDB := testutil.SetupTestDB(ctx, t)
	defer testDB.Teardown(ctx)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	body := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/polls", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler := server.HandleCreatePoll(logger, testDB.Client)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
