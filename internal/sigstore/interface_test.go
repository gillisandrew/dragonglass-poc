package sigstore

import (
	"testing"

	"github.com/gillisandrew/dragonglass-poc/internal/domain"
)

// Compile-time verification that Service implements domain.AttestationService
var _ domain.AttestationService = (*Service)(nil)

// TestAttestationServiceInterface verifies that Sigstore Service properly implements domain.AttestationService
func TestAttestationServiceInterface(t *testing.T) {
	// We can't easily instantiate a Sigstore service without proper dependencies in a unit test,
	// but we can verify the interface assignment at compile time
	var svc *Service
	var _ domain.AttestationService = svc

	// The compile-time check above is the main verification
	// If this test compiles, the interface is correctly implemented
}
