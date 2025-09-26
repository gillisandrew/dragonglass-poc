// ABOUTME: Core domain types for the dragonglass application
// ABOUTME: Contains the fundamental types that define the business domain
package domain

import (
	"time"

	v1 "github.com/in-toto/attestation/go/predicates/provenance/v1"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// Plugin represents the complete metadata for an Obsidian plugin
type Plugin struct {
	// Core plugin information from manifest.json
	ID            string `json:"id"`
	Name          string `json:"name"`
	Version       string `json:"version"`
	MinAppVersion string `json:"minAppVersion,omitempty"`
	Description   string `json:"description,omitempty"`
	Author        string `json:"author,omitempty"`
	AuthorURL     string `json:"authorUrl,omitempty"`
	IsDesktopOnly bool   `json:"isDesktopOnly,omitempty"`
}

// Lockfile represents the plugin lockfile structure
type Lockfile struct {
	SchemaVersion string                       `json:"schemaVersion"`
	VaultName     string                       `json:"vaultName"`
	VaultPath     string                       `json:"vaultPath"`
	Plugins       map[string]PluginEntry       `json:"plugins"`
	Metadata      LockfileMetadata             `json:"metadata"`
	Verification  map[string]VerificationState `json:"verification"`
}

// PluginEntry represents a single plugin entry in the lockfile
type PluginEntry struct {
	Version     string         `json:"version"`
	Registry    string         `json:"registry"`
	Resolved    string         `json:"resolved"`
	Integrity   string         `json:"integrity"`
	Metadata    PluginMetadata `json:"metadata"`
	InstallTime time.Time      `json:"installTime"`
}

// VerificationState tracks the verification status of a plugin
type VerificationState struct {
	Verified         bool      `json:"verified"`
	AttestationValid bool      `json:"attestationValid"`
	SBOMValid        bool      `json:"sbomValid"`
	LastVerified     time.Time `json:"lastVerified"`
	Errors           []string  `json:"errors,omitempty"`
}

// PluginMetadata contains resolved plugin metadata from the registry
type PluginMetadata struct {
	Name          string `json:"name"`
	Version       string `json:"version"`
	Author        string `json:"author"`
	Description   string `json:"description"`
	MinAppVersion string `json:"minAppVersion,omitempty"`
}

// LockfileMetadata contains metadata about the lockfile itself
type LockfileMetadata struct {
	CreatedAt   time.Time `json:"createdAt"`
	LastUpdated time.Time `json:"lastUpdated"`
	Version     string    `json:"version"`
}

// VerificationResult contains comprehensive verification results for all attestation types
type VerificationResult struct {
	Found           bool             `json:"found"`
	Valid           bool             `json:"valid"`
	Errors          []string         `json:"errors"`
	Warnings        []string         `json:"warnings"`
	SLSA            *SLSAResult      `json:"slsa,omitempty"`
	SBOM            *SBOMResult      `json:"sbom,omitempty"`
	AttestationData *AttestationData `json:"attestationData,omitempty"`
}

// SLSAResult contains SLSA provenance verification results
type SLSAResult struct {
	Level           int                 `json:"level"`
	BuilderID       string              `json:"builderID"`
	BuildDefinition *v1.BuildDefinition `json:"buildDefinition,omitempty"`
	RunDetails      *v1.RunDetails      `json:"runDetails,omitempty"`
	TrustedBuilder  bool                `json:"trustedBuilder"`
	WorkflowInputs  map[string]any      `json:"workflowInputs,omitempty"`
}

// SBOMResult contains SBOM verification results
type SBOMResult struct {
	DocumentName    string          `json:"documentName"`
	Packages        []string        `json:"packages"`
	Vulnerabilities []Vulnerability `json:"vulnerabilities,omitempty"`
}

// Vulnerability represents a security vulnerability found in dependencies
type Vulnerability struct {
	ID       string `json:"id"`
	Severity string `json:"severity"`
	Package  string `json:"package"`
	Version  string `json:"version"`
	Summary  string `json:"summary"`
}

// AttestationData contains the raw attestation data and metadata
type AttestationData struct {
	Attestations []ocispec.Descriptor `json:"attestations"`
	Signatures   []ocispec.Descriptor `json:"signatures"`
}

// ValidationError represents a plugin validation error
type ValidationError struct {
	Field   string
	Message string
}

// ValidationResult contains the result of plugin metadata validation
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors"`
}

// Config represents the application configuration
type Config struct {
	// Application settings
	VaultPath string `json:"vaultPath"`
	ConfigDir string `json:"configDir"`
	LogLevel  string `json:"logLevel"`
	LogFile   string `json:"logFile"`

	// Registry settings
	Registry RegistryConfig `json:"registry"`

	// Verification settings
	Verification VerificationConfig `json:"verification"`

	// Output settings
	Output OutputConfig `json:"output"`
}

// VerificationConfig contains verification-related settings
type VerificationConfig struct {
	RequireAttestation  bool   `json:"requireAttestation"`
	TrustedBuilder      string `json:"trustedBuilder"`
	AnnotationNamespace string `json:"annotationNamespace"`
}

// OutputConfig contains output formatting settings
type OutputConfig struct {
	Format string `json:"format"`
	Color  bool   `json:"color"`
	Quiet  bool   `json:"quiet"`
}

// RegistryConfig contains registry connection settings
type RegistryConfig struct {
	DefaultHost string        `json:"defaultHost"`
	Timeout     time.Duration `json:"timeout"`
}

// Service Interfaces - These define the contracts for our dependency packages

// AuthService handles authentication operations
type AuthService interface {
	// Authenticate starts the authentication flow
	Authenticate() error

	// IsAuthenticated checks if user has valid credentials
	IsAuthenticated() bool

	// GetToken returns the current authentication token
	GetToken() (string, error)

	// GetUser returns the authenticated user information
	GetUser() (string, error)

	// Logout clears stored credentials
	Logout() error
}

// RegistryService handles OCI registry operations
type RegistryService interface {
	// Pull downloads a plugin from the registry
	Pull(imageRef string) (*PullResult, error)

	// ValidateAccess checks if we can access the given image reference
	ValidateAccess(imageRef string) error

	// GetManifest retrieves the OCI manifest for an image
	GetManifest(imageRef string) (*ocispec.Manifest, map[string]string, error)
}

// AttestationService handles cryptographic verification
type AttestationService interface {
	// VerifyAttestations verifies SLSA provenance and SBOM attestations
	VerifyAttestations(imageRef string, trustedBuilder string) (*VerificationResult, error)
}

// LockfileService handles lockfile operations
type LockfileService interface {
	// Load reads the lockfile from the vault directory
	Load() (*Lockfile, error)

	// Save writes the lockfile to the vault directory
	Save(lockfile *Lockfile) error

	// AddPlugin adds a plugin entry to the lockfile
	AddPlugin(id string, entry PluginEntry) error

	// RemovePlugin removes a plugin from the lockfile
	RemovePlugin(id string) error

	// UpdateVerification updates verification status for a plugin
	UpdateVerification(id string, verification VerificationState) error
}

// ConfigService handles application configuration
type ConfigService interface {
	// Load reads configuration from the config directory
	Load() (*Config, error)

	// Save writes configuration to the config directory
	Save(config *Config) error

	// GetVaultPath returns the path to the Obsidian vault
	GetVaultPath() (string, error)
}

// PullResult contains the result of pulling an OCI image
type PullResult struct {
	Manifest     ocispec.Manifest
	ManifestData []byte
	Layers       []LayerInfo
	Annotations  map[string]string
	Plugin       *Plugin
}

// LayerInfo contains information about an OCI layer
type LayerInfo struct {
	Descriptor ocispec.Descriptor
	Content    []byte
	SavedPath  string // Path where layer content was saved
}
