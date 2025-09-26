// ABOUTME: Mock authentication provider for testing
// ABOUTME: Provides fake authentication for tests without requiring real credentials
package registry

import (
	"net/http"
	"time"
)

// mockAuthProvider implements AuthProvider for testing
type mockAuthProvider struct {
	token       string
	shouldError bool
}

// NewMockAuthProvider creates a new mock auth provider for testing
func NewMockAuthProvider(token string, shouldError bool) AuthProvider {
	return &mockAuthProvider{
		token:       token,
		shouldError: shouldError,
	}
}

// GetToken returns the mock token or an error if configured to fail
func (m *mockAuthProvider) GetToken() (string, error) {
	if m.shouldError {
		return "", &MockAuthError{Message: "mock authentication error"}
	}
	if m.token == "" {
		return "mock-token", nil
	}
	return m.token, nil
}

// GetHTTPClient returns a basic HTTP client for testing
func (m *mockAuthProvider) GetHTTPClient() (*http.Client, error) {
	if m.shouldError {
		return nil, &MockAuthError{Message: "mock HTTP client error"}
	}
	return &http.Client{
		Timeout: 30 * time.Second,
	}, nil
}

// MockAuthError represents a mock authentication error
type MockAuthError struct {
	Message string
}

func (e *MockAuthError) Error() string {
	return e.Message
}

// IsMockAuthError checks if an error is a mock authentication error
func IsMockAuthError(err error) bool {
	_, ok := err.(*MockAuthError)
	return ok
}
