package domain

import (
	"testing"
)

// This file serves as documentation for the domain interfaces and their implementations
// All interface compliance tests are handled in their respective packages

// TestDomainInterfaces documents the domain interface contracts
func TestDomainInterfaces(t *testing.T) {
	t.Run("AuthService", func(t *testing.T) {
		t.Log("AuthService interface provides authentication methods:")
		t.Log("  - Authenticate() error")
		t.Log("  - IsAuthenticated() bool")
		t.Log("  - GetToken() (string, error)")
		t.Log("  - GetUser() (string, error)")
		t.Log("  - GetCredential() (*Credential, error)")
		t.Log("  - Logout() error")
		t.Log("Implementation: internal/github/service.go")
	})

	t.Run("RegistryService", func(t *testing.T) {
		t.Log("RegistryService interface provides OCI registry operations:")
		t.Log("  - Pull(imageRef string) (*PullResult, error)")
		t.Log("  - ValidateAccess(imageRef string) error")
		t.Log("  - GetManifest(imageRef string) (*ocispec.Manifest, map[string]string, error)")
		t.Log("Implementation: internal/oras/service.go")
	})

	t.Run("AttestationService", func(t *testing.T) {
		t.Log("AttestationService interface provides cryptographic verification:")
		t.Log("  - VerifyAttestations(imageRef, trustedBuilder string) (*VerificationResult, error)")
		t.Log("Implementation: internal/sigstore/service.go")
	})

	t.Run("AdditionalInterfaces", func(t *testing.T) {
		t.Log("oras.AuthProvider interface (implemented by GitHub service):")
		t.Log("  - GetToken() (string, error)")
		t.Log("This allows GitHub service to provide authentication for ORAS operations")
	})
}
