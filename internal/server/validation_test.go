package server

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  string
	}{
		{
			name:     "valid username",
			username: "johndoe",
			wantErr:  "",
		},
		{
			name:     "valid username with underscore",
			username: "john_doe",
			wantErr:  "",
		},
		{
			name:     "valid username with hyphen",
			username: "john-doe",
			wantErr:  "",
		},
		{
			name:     "valid username with numbers",
			username: "john123",
			wantErr:  "",
		},
		{
			name:     "empty username",
			username: "",
			wantErr:  "username is required",
		},
		{
			name:     "whitespace only",
			username: "   ",
			wantErr:  "username is required",
		},
		{
			name:     "too short",
			username: "ab",
			wantErr:  "username must be at least 3 characters",
		},
		{
			name:     "too long",
			username: strings.Repeat("a", 33),
			wantErr:  "username must be at most 32 characters",
		},
		{
			name:     "starts with number",
			username: "1user",
			wantErr:  "username must start with a letter",
		},
		{
			name:     "starts with underscore",
			username: "_user",
			wantErr:  "username must start with a letter",
		},
		{
			name:     "contains invalid character @",
			username: "user@name",
			wantErr:  "username can only contain letters, numbers, underscores, and hyphens",
		},
		{
			name:     "contains space",
			username: "user name",
			wantErr:  "username can only contain letters, numbers, underscores, and hyphens",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateUsername(tt.username)
			assert.Equal(t, tt.wantErr, got)
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr string
	}{
		{
			name:    "valid email",
			email:   "test@example.com",
			wantErr: "",
		},
		{
			name:    "valid email with subdomain",
			email:   "test@mail.example.com",
			wantErr: "",
		},
		{
			name:    "valid email with plus",
			email:   "test+tag@example.com",
			wantErr: "",
		},
		{
			name:    "empty email",
			email:   "",
			wantErr: "email is required",
		},
		{
			name:    "whitespace only",
			email:   "   ",
			wantErr: "email is required",
		},
		{
			name:    "missing @",
			email:   "testexample.com",
			wantErr: "invalid email format",
		},
		{
			name:    "missing domain",
			email:   "test@",
			wantErr: "invalid email format",
		},
		{
			name:    "missing TLD",
			email:   "test@example",
			wantErr: "invalid email format",
		},
		{
			name:    "invalid format",
			email:   "not-an-email",
			wantErr: "invalid email format",
		},
		{
			name:    "too long",
			email:   strings.Repeat("a", 250) + "@example.com",
			wantErr: "email address is too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateEmail(tt.email)
			assert.Equal(t, tt.wantErr, got)
		})
	}
}

func TestValidatePollTitle(t *testing.T) {
	tests := []struct {
		name    string
		title   string
		wantErr string
	}{
		{
			name:    "valid title",
			title:   "What's your favorite color?",
			wantErr: "",
		},
		{
			name:    "empty title",
			title:   "",
			wantErr: "title is required",
		},
		{
			name:    "whitespace only",
			title:   "   ",
			wantErr: "title is required",
		},
		{
			name:    "too long",
			title:   strings.Repeat("a", 257),
			wantErr: "title must be at most 256 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidatePollTitle(tt.title)
			assert.Equal(t, tt.wantErr, got)
		})
	}
}

func TestValidatePollOptions(t *testing.T) {
	tests := []struct {
		name    string
		options []string
		wantErr string
	}{
		{
			name:    "valid options",
			options: []string{"Option 1", "Option 2"},
			wantErr: "",
		},
		{
			name:    "valid multiple options",
			options: []string{"A", "B", "C", "D"},
			wantErr: "",
		},
		{
			name:    "too few options",
			options: []string{"Only One"},
			wantErr: "at least 2 options are required",
		},
		{
			name:    "empty options",
			options: []string{},
			wantErr: "at least 2 options are required",
		},
		{
			name:    "nil options",
			options: nil,
			wantErr: "at least 2 options are required",
		},
		{
			name:    "too many options",
			options: make([]string, 21),
			wantErr: "at most 20 options are allowed",
		},
		{
			name:    "empty option text",
			options: []string{"Option 1", ""},
			wantErr: "option cannot be empty",
		},
		{
			name:    "whitespace option",
			options: []string{"Option 1", "   "},
			wantErr: "option cannot be empty",
		},
		{
			name:    "option too long",
			options: []string{"Option 1", strings.Repeat("a", 257)},
			wantErr: "option text must be at most 256 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy to avoid modifying test data
			opts := make([]string, len(tt.options))
			copy(opts, tt.options)
			got := ValidatePollOptions(opts)
			assert.Equal(t, tt.wantErr, got)
		})
	}
}
