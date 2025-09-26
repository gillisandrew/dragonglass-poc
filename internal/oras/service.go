// ABOUTME: ORAS adapter - wraps OCI registry operations using ORAS library
// ABOUTME: Implements domain.RegistryService interface for OCI artifact operations
package oras

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"

	"github.com/gillisandrew/dragonglass-poc/internal/domain"
)

// Service implements domain.RegistryService using ORAS
type Service struct {
	registry   *remote.Registry
	httpClient *http.Client
	timeout    time.Duration
}

// AuthProvider interface for dependency injection
type AuthProvider interface {
	GetToken() (string, error)
}

// NewService creates a new ORAS registry service
func NewService(registryHost string, authProvider AuthProvider) (*Service, error) {
	// Get authentication token
	token, err := authProvider.GetToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get authentication token: %w", err)
	}

	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create ORAS registry client
	registry, err := remote.NewRegistry(registryHost)
	if err != nil {
		return nil, fmt.Errorf("failed to create registry client: %w", err)
	}

	// Configure ORAS authentication
	registry.Client = &auth.Client{
		Client: retry.DefaultClient,
		Cache:  auth.NewCache(),
		Credential: auth.StaticCredential(registryHost, auth.Credential{
			Username: "token",
			Password: token,
		}),
	}

	return &Service{
		registry:   registry,
		httpClient: httpClient,
		timeout:    30 * time.Second,
	}, nil
}

// Pull implements domain.RegistryService.Pull
func (s *Service) Pull(imageRef string) (*domain.PullResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	// Parse image reference
	registryHost, repository, tag, err := parseImageReference(imageRef)
	if err != nil {
		return nil, fmt.Errorf("invalid image reference: %w", err)
	}

	// Create repository client
	repo, err := s.registry.Repository(ctx, repository)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository client: %w", err)
	}

	// Setup authentication for repository
	if remoteRepo, ok := repo.(*remote.Repository); ok {
		s.setupRepositoryAuth(remoteRepo, registryHost)
	}

	// Get manifest
	manifestDesc, manifest, err := s.getManifest(ctx, repo, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to get manifest: %w", err)
	}

	// Pull layers
	layers, err := s.pullLayers(ctx, repo, manifest.Layers)
	if err != nil {
		return nil, fmt.Errorf("failed to pull layers: %w", err)
	}

	// Extract annotations
	annotations := make(map[string]string)
	for k, v := range manifest.Annotations {
		annotations[k] = v
	}
	for k, v := range manifestDesc.Annotations {
		annotations[k] = v
	}

	result := &domain.PullResult{
		Manifest:     *manifest,
		ManifestData: nil, // TODO: Add manifest data if needed
		Layers:       layers,
		Annotations:  annotations,
		Plugin:       nil, // Will be populated by plugin parsing
	}

	return result, nil
}

// ValidateAccess implements domain.RegistryService.ValidateAccess
func (s *Service) ValidateAccess(imageRef string) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	// Parse image reference
	_, repository, tag, err := parseImageReference(imageRef)
	if err != nil {
		return fmt.Errorf("invalid image reference: %w", err)
	}

	// Create repository client
	repo, err := s.registry.Repository(ctx, repository)
	if err != nil {
		return fmt.Errorf("failed to create repository client: %w", err)
	}

	// Try to resolve the manifest to test access
	_, err = repo.Resolve(ctx, tag)
	if err != nil {
		return fmt.Errorf("access validation failed for %s: %w", imageRef, err)
	}

	return nil
}

// GetManifest implements domain.RegistryService.GetManifest
func (s *Service) GetManifest(imageRef string) (*ocispec.Manifest, map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	// Parse image reference
	registryHost, repository, tag, err := parseImageReference(imageRef)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid image reference: %w", err)
	}

	// Create repository client
	repo, err := s.registry.Repository(ctx, repository)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create repository client: %w", err)
	}

	// Setup authentication
	if remoteRepo, ok := repo.(*remote.Repository); ok {
		s.setupRepositoryAuth(remoteRepo, registryHost)
	}

	// Get manifest
	manifestDesc, manifest, err := s.getManifest(ctx, repo, tag)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get manifest: %w", err)
	}

	// Combine annotations
	annotations := make(map[string]string)
	for k, v := range manifest.Annotations {
		annotations[k] = v
	}
	for k, v := range manifestDesc.Annotations {
		annotations[k] = v
	}

	return manifest, annotations, nil
}

// Helper methods

func (s *Service) setupRepositoryAuth(repo *remote.Repository, registryHost string) {
	// The repository inherits auth from the registry, but we can override if needed
	// For now, the registry-level auth should be sufficient
}

func (s *Service) getManifest(ctx context.Context, repo registry.Repository, tag string) (ocispec.Descriptor, *ocispec.Manifest, error) {
	// Resolve the tag to get the manifest descriptor
	manifestDesc, err := repo.Resolve(ctx, tag)
	if err != nil {
		return ocispec.Descriptor{}, nil, fmt.Errorf("failed to resolve tag: %w", err)
	}

	// Fetch the manifest content
	manifestReader, err := repo.Fetch(ctx, manifestDesc)
	if err != nil {
		return ocispec.Descriptor{}, nil, fmt.Errorf("failed to fetch manifest: %w", err)
	}
	defer manifestReader.Close()

	// Parse the manifest
	var manifest ocispec.Manifest
	if err := json.NewDecoder(manifestReader).Decode(&manifest); err != nil {
		return ocispec.Descriptor{}, nil, fmt.Errorf("failed to decode manifest: %w", err)
	}

	return manifestDesc, &manifest, nil
}

func (s *Service) pullLayers(ctx context.Context, repo registry.Repository, layerDescs []ocispec.Descriptor) ([]domain.LayerInfo, error) {
	var layers []domain.LayerInfo

	for _, desc := range layerDescs {
		// Fetch layer content
		layerReader, err := repo.Fetch(ctx, desc)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch layer %s: %w", desc.Digest, err)
		}

		// Read layer content
		content, err := io.ReadAll(layerReader)
		layerReader.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read layer content: %w", err)
		}

		layer := domain.LayerInfo{
			Descriptor: desc,
			Content:    content,
			SavedPath:  "", // TODO: Add layer saving if needed
		}

		layers = append(layers, layer)
	}

	return layers, nil
}

// parseImageReference parses an OCI image reference into components
func parseImageReference(imageRef string) (registryHost, repository, tag string, err error) {
	// Simple parsing - in production would use a proper OCI reference parser
	parts := strings.Split(imageRef, "/")
	if len(parts) < 2 {
		return "", "", "", fmt.Errorf("invalid image reference format")
	}

	registryHost = parts[0]

	repoAndTag := strings.Join(parts[1:], "/")
	if strings.Contains(repoAndTag, ":") {
		repoParts := strings.Split(repoAndTag, ":")
		repository = repoParts[0]
		tag = repoParts[1]
	} else if strings.Contains(repoAndTag, "@") {
		repoParts := strings.Split(repoAndTag, "@")
		repository = repoParts[0]
		tag = repoParts[1] // digest
	} else {
		repository = repoAndTag
		tag = "latest"
	}

	return registryHost, repository, tag, nil
}
