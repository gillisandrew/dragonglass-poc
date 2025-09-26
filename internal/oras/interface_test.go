package oras

import (
	"testing"

	"github.com/gillisandrew/dragonglass-poc/internal/domain"
)

// Compile-time verification that Service implements domain.RegistryService
var _ domain.RegistryService = (*Service)(nil)

// TestRegistryServiceInterface verifies that ORAS Service properly implements domain.RegistryService
func TestRegistryServiceInterface(t *testing.T) {
	// We can't easily instantiate an ORAS service without proper dependencies in a unit test,
	// but we can verify the interface assignment at compile time
	var svc *Service
	var _ domain.RegistryService = svc

	// The compile-time check above is the main verification
	// If this test compiles, the interface is correctly implemented
}
