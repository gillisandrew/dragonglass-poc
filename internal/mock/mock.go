// ABOUTME: Shared mock implementations for testing across all packages
// ABOUTME: Centralizes test doubles to avoid duplication and improve test maintainability
package mock

import (
	"net/http"
	"time"

	"github.com/gillisandrew/dragonglass-poc/internal/domain"
)

// AuthProvider provides mock authentication for testing
type AuthProvider struct {
	token       string
	shouldError bool
}

// NewAuthProvider creates a new mock auth provider for testing
func NewAuthProvider(token string, shouldError bool) *AuthProvider {
	return &AuthProvider{
		token:       token,
		shouldError: shouldError,
	}
}

// GetToken returns the mock token or an error if configured to fail
func (m *AuthProvider) GetToken() (string, error) {
	if m.shouldError {
		return "", &AuthError{Message: "mock authentication error"}
	}
	if m.token == "" {
		return "mock-token", nil
	}
	return m.token, nil
}

// GetHTTPClient returns a basic HTTP client for testing
func (m *AuthProvider) GetHTTPClient() (*http.Client, error) {
	if m.shouldError {
		return nil, &AuthError{Message: "mock HTTP client error"}
	}
	return &http.Client{
		Timeout: 30 * time.Second, // DefaultTimeout equivalent
	}, nil
}

// AuthError represents a mock authentication error
type AuthError struct {
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
}

// IsAuthError checks if an error is a mock authentication error
func IsAuthError(err error) bool {
	_, ok := err.(*AuthError)
	return ok
}

// Repository provides a mock OCI repository for testing
type Repository struct {
	shouldError  bool
	manifestData []byte
	layerData    map[string][]byte
}

// NewRepository creates a new mock repository
func NewRepository(shouldError bool) *Repository {
	return &Repository{
		shouldError: shouldError,
		layerData:   make(map[string][]byte),
	}
}

// SetManifest sets mock manifest data
func (r *Repository) SetManifest(data []byte) {
	r.manifestData = data
}

// SetLayerData sets mock layer data for a given digest
func (r *Repository) SetLayerData(digest string, data []byte) {
	r.layerData[digest] = data
}

// Plugin creates a mock plugin for testing
func Plugin() domain.Plugin {
	return domain.Plugin{
		ID:            "test-plugin",
		Name:          "Test Plugin",
		Version:       "1.0.0",
		MinAppVersion: "0.15.0",
		Description:   "A test plugin for testing",
		Author:        "Test Author",
		AuthorURL:     "https://example.com",
		IsDesktopOnly: false,
	}
}

// Lockfile creates a mock lockfile for testing
func Lockfile() domain.Lockfile {
	now := time.Now()
	return domain.Lockfile{
		SchemaVersion: "1.0.0",
		VaultName:     "Test Vault",
		VaultPath:     "/path/to/vault",
		Plugins: map[string]domain.PluginEntry{
			"test-plugin": {
				Version:   "1.0.0",
				Registry:  "ghcr.io",
				Resolved:  "ghcr.io/test/plugin@sha256:abc123",
				Integrity: "sha256:abc123",
				Metadata: domain.PluginMetadata{
					Name:        "Test Plugin",
					Version:     "1.0.0",
					Author:      "Test Author",
					Description: "A test plugin",
				},
				InstallTime: now,
			},
		},
		Metadata: domain.LockfileMetadata{
			CreatedAt:   now,
			LastUpdated: now,
			Version:     "1.0.0",
		},
		Verification: map[string]domain.VerificationState{
			"test-plugin": {
				Verified:         true,
				AttestationValid: true,
				SBOMValid:        true,
				LastVerified:     now,
				Errors:           nil,
			},
		},
	}
}

// Config creates a mock config for testing
func Config() domain.Config {
	return domain.Config{
		VaultPath: "/path/to/vault",
		ConfigDir: "/path/to/config",
		LogLevel:  "info",
		LogFile:   "/path/to/log",
		Registry: domain.RegistryConfig{
			DefaultHost: "ghcr.io",
			Timeout:     30 * time.Second,
		},
		Verification: domain.VerificationConfig{
			RequireAttestation:  true,
			TrustedBuilder:      "https://github.com/test/repo/.github/workflows/build.yml@refs/heads/main",
			AnnotationNamespace: "md.obsidian.plugin.v0",
		},
		Output: domain.OutputConfig{
			Format: "json",
			Color:  true,
			Quiet:  false,
		},
	}
}

// VerificationResult creates a mock verification result for testing
func VerificationResult() domain.VerificationResult {
	return domain.VerificationResult{
		Found:    true,
		Valid:    true,
		Errors:   nil,
		Warnings: nil,
		SLSA: &domain.SLSAResult{
			Level:          3,
			BuilderID:      "https://github.com/actions/runner",
			TrustedBuilder: true,
		},
		SBOM: &domain.SBOMResult{
			DocumentName: "test-sbom",
			Packages:     []string{"package1", "package2"},
		},
	}
}
