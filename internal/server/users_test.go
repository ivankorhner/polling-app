//go:build integration

package server_test

import (
	"bytes"
	"context"
	"encoding/json"
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

func TestHandleRegisterUser_Success(t *testing.T) {
	ctx := context.Background()
	testDB := testutil.SetupTestDB(ctx, t)
	defer testDB.Teardown(ctx)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	body := `{"username": "testuser", "email": "test@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler := server.HandleRegisterUser(logger, testDB.Client)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var result server.UserResponse
	err := json.Unmarshal(rec.Body.Bytes(), &result)
	require.NoError(t, err)

	assert.Equal(t, "testuser", result.Username)
	assert.Equal(t, "test@example.com", result.Email)
	assert.NotZero(t, result.ID)
	assert.NotZero(t, result.CreatedAt)
}

func TestHandleRegisterUser_Validation(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantError  string
	}{
		{
			name:       "missing username",
			body:       `{"email": "test@example.com"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "username is required",
		},
		{
			name:       "missing email",
			body:       `{"username": "testuser"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "email is required",
		},
		{
			name:       "invalid JSON",
			body:       `{invalid json}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid request body",
		},
		{
			name:       "username too short",
			body:       `{"username": "ab", "email": "test@example.com"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "username must be at least 3 characters",
		},
		{
			name:       "username starts with number",
			body:       `{"username": "1user", "email": "test@example.com"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "username must start with a letter",
		},
		{
			name:       "username with invalid characters",
			body:       `{"username": "user@name", "email": "test@example.com"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "username can only contain letters, numbers, underscores, and hyphens",
		},
		{
			name:       "invalid email format",
			body:       `{"username": "testuser", "email": "not-an-email"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid email format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			testDB := testutil.SetupTestDB(ctx, t)
			defer testDB.Teardown(ctx)

			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler := server.HandleRegisterUser(logger, testDB.Client)
			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)

			var errResp server.ErrorResponse
			err := json.Unmarshal(rec.Body.Bytes(), &errResp)
			require.NoError(t, err)
			assert.Equal(t, tt.wantError, errResp.Error)
		})
	}
}

func TestHandleRegisterUser_Conflict(t *testing.T) {
	tests := []struct {
		name          string
		existingUser  map[string]string
		newUser       map[string]string
		wantStatus    int
	}{
		{
			name: "duplicate username",
			existingUser: map[string]string{
				"username": "existinguser",
				"email":    "existing@example.com",
			},
			newUser: map[string]string{
				"username": "existinguser",
				"email":    "new@example.com",
			},
			wantStatus: http.StatusConflict,
		},
		{
			name: "duplicate email",
			existingUser: map[string]string{
				"username": "existinguser",
				"email":    "existing@example.com",
			},
			newUser: map[string]string{
				"username": "newuser",
				"email":    "existing@example.com",
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			testDB := testutil.SetupTestDB(ctx, t)
			defer testDB.Teardown(ctx)

			// Create existing user
			_, err := testDB.Client.User.Create().
				SetUsername(tt.existingUser["username"]).
				SetEmail(tt.existingUser["email"]).
				Save(ctx)
			require.NoError(t, err)

			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

			body, _ := json.Marshal(tt.newUser)
			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler := server.HandleRegisterUser(logger, testDB.Client)
			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}
