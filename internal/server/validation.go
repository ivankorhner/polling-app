package server

import (
	"regexp"
	"strings"
	"unicode"
)

const (
	// UsernameMinLength is the minimum length for usernames
	UsernameMinLength = 3
	// UsernameMaxLength is the maximum length for usernames
	UsernameMaxLength = 32
)

// emailRegex is a simple but effective email validation pattern
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// ValidateUsername validates a username and returns an error message if invalid
func ValidateUsername(username string) string {
	username = strings.TrimSpace(username)

	if username == "" {
		return "username is required"
	}

	if len(username) < UsernameMinLength {
		return "username must be at least 3 characters"
	}

	if len(username) > UsernameMaxLength {
		return "username must be at most 32 characters"
	}

	// Check that username contains only allowed characters (alphanumeric, underscore, hyphen)
	for _, r := range username {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '-' {
			return "username can only contain letters, numbers, underscores, and hyphens"
		}
	}

	// Username must start with a letter
	if !unicode.IsLetter(rune(username[0])) {
		return "username must start with a letter"
	}

	return ""
}

// ValidateEmail validates an email address and returns an error message if invalid
func ValidateEmail(email string) string {
	email = strings.TrimSpace(email)

	if email == "" {
		return "email is required"
	}

	if len(email) > 254 { // RFC 5321
		return "email address is too long"
	}

	if !emailRegex.MatchString(email) {
		return "invalid email format"
	}

	return ""
}

// ValidatePollTitle validates a poll title and returns an error message if invalid
func ValidatePollTitle(title string) string {
	title = strings.TrimSpace(title)

	if title == "" {
		return "title is required"
	}

	if len(title) > 256 {
		return "title must be at most 256 characters"
	}

	return ""
}

// ValidatePollOptions validates poll options and returns an error message if invalid
func ValidatePollOptions(options []string) string {
	if len(options) < 2 {
		return "at least 2 options are required"
	}

	if len(options) > 20 {
		return "at most 20 options are allowed"
	}

	for i, opt := range options {
		opt = strings.TrimSpace(opt)
		if opt == "" {
			return "option cannot be empty"
		}
		if len(opt) > 256 {
			return "option text must be at most 256 characters"
		}
		options[i] = opt // Normalize trimmed value
	}

	return ""
}
