package github

import (
	"testing"

	"github.com/gillisandrew/dragonglass-poc/internal/domain"
	"github.com/gillisandrew/dragonglass-poc/internal/oras"
)

// Compile-time verification that Service implements domain.AuthService
var _ domain.AuthService = (*Service)(nil)

// Compile-time verification that Service implements oras.AuthProvider
// This ensures GitHub service can provide authentication for ORAS operations
var _ oras.AuthProvider = (*Service)(nil)

// TestAuthServiceInterface verifies that GitHub Service properly implements domain.AuthService
func TestAuthServiceInterface(t *testing.T) {
	svc := NewService()

	// Test that we can assign to interface types
	var authService domain.AuthService = svc
	var authProvider oras.AuthProvider = svc

	if authService == nil {
		t.Error("Service should implement domain.AuthService")
	}
	if authProvider == nil {
		t.Error("Service should implement oras.AuthProvider")
	}
}
