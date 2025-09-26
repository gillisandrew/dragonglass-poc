package main

import (
	"testing"

	"github.com/gillisandrew/dragonglass-poc/internal/domain"
	"github.com/gillisandrew/dragonglass-poc/internal/github"
	"github.com/gillisandrew/dragonglass-poc/internal/oras"
	"github.com/gillisandrew/dragonglass-poc/internal/sigstore"
)

// Compile-time verification that all services implement their domain interfaces
var (
	_ domain.AuthService        = (*github.Service)(nil)
	_ domain.RegistryService    = (*oras.Service)(nil)
	_ domain.AttestationService = (*sigstore.Service)(nil)
	_ oras.AuthProvider         = (*github.Service)(nil) // GitHub service provides auth for ORAS
)

// TestInterfaceCompliance verifies that all services implement their interfaces correctly
func TestInterfaceCompliance(t *testing.T) {
	t.Run("AuthService", func(t *testing.T) {
		var svc domain.AuthService = github.NewService()
		if svc == nil {
			t.Error("GitHub service should implement AuthService")
		}
	})

	t.Run("RegistryService", func(t *testing.T) {
		// We can't easily create an ORAS service without dependencies in a test,
		// but the compile-time check above ensures interface compliance
		var svc *oras.Service
		var _ domain.RegistryService = svc
	})

	t.Run("AttestationService", func(t *testing.T) {
		// We can't easily create a Sigstore service without dependencies in a test,
		// but the compile-time check above ensures interface compliance
		var svc *sigstore.Service
		var _ domain.AttestationService = svc
	})

	t.Run("AuthProvider", func(t *testing.T) {
		var svc oras.AuthProvider = github.NewService()
		if svc == nil {
			t.Error("GitHub service should implement AuthProvider")
		}
	})
}
