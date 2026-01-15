//go:build integration

package handlers_test

import (
	"bytes"
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

func TestHandleRegisterUser_Success(t *testing.T) {
	ctx := context.Background()
	testDB := testutil.SetupTestDB(ctx, t)
	defer testDB.Teardown(ctx)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	body := `{"username": "testuser", "email": "test@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler := handlers.HandleRegisterUser(logger, testDB.Client)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var result handlers.UserResponse
	err := json.Unmarshal(rec.Body.Bytes(), &result)
	require.NoError(t, err)

	assert.Equal(t, "testuser", result.Username)
	assert.Equal(t, "test@example.com", result.Email)
	assert.NotZero(t, result.ID)
	assert.NotZero(t, result.CreatedAt)
}

func TestHandleRegisterUser_DuplicateUsername(t *testing.T) {
	ctx := context.Background()
	testDB := testutil.SetupTestDB(ctx, t)
	defer testDB.Teardown(ctx)

	// Create existing user
	_, err := testDB.Client.User.Create().
		SetUsername("existinguser").
		SetEmail("existing@example.com").
		Save(ctx)
	require.NoError(t, err)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	body := `{"username": "existinguser", "email": "new@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler := handlers.HandleRegisterUser(logger, testDB.Client)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestHandleRegisterUser_DuplicateEmail(t *testing.T) {
	ctx := context.Background()
	testDB := testutil.SetupTestDB(ctx, t)
	defer testDB.Teardown(ctx)

	// Create existing user
	_, err := testDB.Client.User.Create().
		SetUsername("existinguser").
		SetEmail("existing@example.com").
		Save(ctx)
	require.NoError(t, err)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	body := `{"username": "newuser", "email": "existing@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler := handlers.HandleRegisterUser(logger, testDB.Client)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestHandleRegisterUser_MissingUsername(t *testing.T) {
	ctx := context.Background()
	testDB := testutil.SetupTestDB(ctx, t)
	defer testDB.Teardown(ctx)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	body := `{"email": "test@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler := handlers.HandleRegisterUser(logger, testDB.Client)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandleRegisterUser_MissingEmail(t *testing.T) {
	ctx := context.Background()
	testDB := testutil.SetupTestDB(ctx, t)
	defer testDB.Teardown(ctx)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	body := `{"username": "testuser"}`
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler := handlers.HandleRegisterUser(logger, testDB.Client)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandleRegisterUser_InvalidJSON(t *testing.T) {
	ctx := context.Background()
	testDB := testutil.SetupTestDB(ctx, t)
	defer testDB.Teardown(ctx)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	body := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler := handlers.HandleRegisterUser(logger, testDB.Client)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
