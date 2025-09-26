// ABOUTME: Verify command for checking plugin security without installation
// ABOUTME: Performs verification checks and displays results without installing
package verify

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/gillisandrew/dragonglass-poc/internal/attestation"
	"github.com/gillisandrew/dragonglass-poc/internal/auth"
	"github.com/gillisandrew/dragonglass-poc/internal/cmd"
	"github.com/gillisandrew/dragonglass-poc/internal/config"
	"github.com/gillisandrew/dragonglass-poc/internal/plugin"
	"github.com/gillisandrew/dragonglass-poc/internal/registry"
)

func NewVerifyCommand(ctx *cmd.CommandContext) *cobra.Command {
	return &cobra.Command{
		Use:   "verify [OCI_IMAGE_REFERENCE]",
		Short: "Verify a plugin without installing it",
		Long: `Verify an Obsidian plugin's provenance and security without installation.
This command downloads and verifies SLSA attestations, SBOM data, and
vulnerability information, then displays the results.

Example:
  dragonglass verify ghcr.io/owner/repo:plugin-name-v1.0.0`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			imageRef := args[0]
			ctx.Logger.Info("Verifying plugin", ctx.Logger.Args("imageRef", imageRef))

			if err := verifyPlugin(imageRef, ctx); err != nil {
				ctx.Logger.Error("Verification failed", ctx.Logger.Args("error", err))
				os.Exit(1)
			}

			ctx.Logger.Info("Plugin verification completed successfully")
		},
	}
}

func verifyPlugin(imageRef string, ctx *cmd.CommandContext) error {
	ctx.Logger.Debug("Creating registry client")

	// Load configuration
	configOpts := config.DefaultConfigOpts()
	if ctx.ConfigPath != "" {
		configOpts = configOpts.WithConfigPath(ctx.ConfigPath)
	}
	configManager := config.NewConfigManager(configOpts)
	cfg, _, err := configManager.LoadConfig()
	if err != nil {
		ctx.Logger.Warn("Failed to load configuration, using defaults", ctx.Logger.Args("error", err))
		cfg = config.DefaultConfig()
	}

	ctx.Logger.Debug("Verification configuration", ctx.Logger.Args("strict", cfg.Verification.StrictMode))

	// Debug: Log token availability
	if ctx.GitHubToken != "" {
		ctx.Logger.Debug("GitHub token provided via flag", ctx.Logger.Args("tokenLength", len(ctx.GitHubToken)))
	} else {
		ctx.Logger.Debug("No GitHub token provided via flag, will attempt to use stored credentials")
	}

	// Configure registry client
	registryOpts := registry.DefaultRegistryOpts()
	if cfg.Registry.DefaultRegistry != "" {
		registryOpts = registryOpts.WithRegistryHost(cfg.Registry.DefaultRegistry)
	}

	// Configure plugin options for registry client
	pluginOpts := &plugin.PluginOpts{
		AnnotationNamespace: ctx.AnnotationNamespace,
	}
	if ctx.TrustedBuilder != "" {
		pluginOpts = pluginOpts.WithTrustedWorkflowSigner(ctx.TrustedBuilder)
	}
	if cfg.Verification.StrictMode {
		pluginOpts = pluginOpts.WithStrictValidation(true)
	}
	registryOpts = registryOpts.WithPluginOpts(pluginOpts)

	// Configure auth - always try to provide an auth provider
	var authProvider *auth.AuthClient
	if ctx.GitHubToken != "" {
		ctx.Logger.Debug("Configuring registry with provided GitHub token")
		authOpts := auth.DefaultAuthOpts().WithToken(ctx.GitHubToken)
		authProvider = auth.NewAuthClient(authOpts)
	} else {
		// Try to get token from environment variables as fallback (useful in CI)
		if ghToken := os.Getenv("GITHUB_TOKEN"); ghToken != "" {
			ctx.Logger.Debug("Using GITHUB_TOKEN environment variable")
			authOpts := auth.DefaultAuthOpts().WithToken(ghToken)
			authProvider = auth.NewAuthClient(authOpts)
		} else if ghToken := os.Getenv("GH_TOKEN"); ghToken != "" {
			ctx.Logger.Debug("Using GH_TOKEN environment variable")
			authOpts := auth.DefaultAuthOpts().WithToken(ghToken)
			authProvider = auth.NewAuthClient(authOpts)
		} else {
			ctx.Logger.Debug("Using default registry authentication (stored credentials)")
			// Default auth adapter - will try stored credentials
		}
	}

	if authProvider != nil {
		registryOpts = registryOpts.WithAuthProvider(authProvider)
	}

	// Create registry client
	client, err := registry.NewClient(registryOpts)
	if err != nil {
		return fmt.Errorf("failed to create registry client: %w", err)
	}

	// Create context with timeout
	opCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ctx.Logger.Debug("Fetching manifest from registry")

	// Get manifest and annotations
	manifest, annotations, _, err := client.GetManifest(opCtx, imageRef)
	if err != nil {
		return fmt.Errorf("failed to fetch manifest: %w", err)
	}

	ctx.Logger.Info("Manifest retrieved successfully",
		ctx.Logger.Args(
			"configSize", manifest.Config.Size,
			"layers", len(manifest.Layers),
			"annotations", len(annotations),
		))

	// Parse plugin metadata from annotations
	ctx.Logger.Debug("Parsing plugin metadata")

	// Use the same plugin options that were configured for the registry client
	parser := plugin.NewManifestParser(pluginOpts)
	pluginMetadata, err := parser.ParseMetadata(manifest, annotations)
	if err != nil {
		return fmt.Errorf("failed to parse plugin metadata: %w", err)
	}

	// Display plugin information
	ctx.Logger.Info("Plugin Information",
		ctx.Logger.Args(
			"name", pluginMetadata.Name,
			"version", pluginMetadata.Version,
			"author", pluginMetadata.Author,
			"description", pluginMetadata.Description,
			"minAppVersion", pluginMetadata.MinAppVersion,
			"authorURL", pluginMetadata.AuthorURL,
		))

	// Show raw annotations for debugging
	if len(annotations) > 0 {
		annotationMap := make(map[string]any)
		for k, v := range annotations {
			annotationMap[k] = v
		}
		ctx.Logger.Debug("Raw annotations", ctx.Logger.ArgsFromMap(annotationMap))
	}

	// Validate metadata
	ctx.Logger.Debug("Validating plugin metadata")
	validation := parser.ValidateMetadata(pluginMetadata)

	if !validation.Valid {
		for _, err := range validation.Errors {
			ctx.Logger.Error("Metadata validation error", ctx.Logger.Args("error", err))
		}
		if !cfg.Verification.StrictMode {
			ctx.Logger.Warn("Continuing in non-strict mode despite validation errors")
		} else {
			return fmt.Errorf("metadata validation failed in strict mode")
		}
	}

	if len(validation.Warnings) > 0 {
		for _, warning := range validation.Warnings {
			ctx.Logger.Warn("Metadata validation warning", ctx.Logger.Args("warning", warning))
		}
	}

	ctx.Logger.Info("Basic verification completed")

	// Get GitHub token for attestation verification
	ctx.Logger.Debug("Getting authentication token")
	token, err := auth.GetToken()
	if err != nil {
		return fmt.Errorf("failed to get authentication token for attestation verification: %w", err)
	}

	// Verify all attestations (SLSA, SBOM, etc.)
	ctx.Logger.Debug("Verifying attestations (SLSA, SBOM, etc.)")
	verifier, err := attestation.NewAttestationVerifier(token, ctx.TrustedBuilder)
	if err != nil {
		return fmt.Errorf("failed to create attestation verifier: %w", err)
	}

	attestationResult, err := verifier.VerifyAttestations(opCtx, imageRef)
	if err != nil {
		return fmt.Errorf("failed to verify attestations: %w", err)
	}

	// Display debug warnings first
	if len(attestationResult.Warnings) > 0 {
		for _, warning := range attestationResult.Warnings {
			ctx.Logger.Debug("Attestation verification debug info", ctx.Logger.Args("info", warning))
		}
	}

	// Display attestation verification results
	ctx.Logger.Info("Attestation verification results", ctx.Logger.Args("found", attestationResult.Found, "valid", attestationResult.Valid))

	// Check if attestation verification should block installation
	if cfg.Verification.StrictMode && (!attestationResult.Found || !attestationResult.Valid) {
		if !attestationResult.Found {
			return fmt.Errorf("attestations not found (required in strict mode)")
		}
		if !attestationResult.Valid {
			return fmt.Errorf("attestation verification failed (required in strict mode)")
		}
	}

	// Additional SBOM-specific security checks
	if attestationResult.SBOM != nil && len(attestationResult.SBOM.Vulnerabilities) > 0 {
		highSeverityVulns := 0
		for _, vuln := range attestationResult.SBOM.Vulnerabilities {
			if vuln.Severity == "HIGH" || vuln.Severity == "CRITICAL" {
				highSeverityVulns++
			}
		}

		if highSeverityVulns > 0 {
			ctx.Logger.Warn("High/critical severity vulnerabilities found", ctx.Logger.Args("count", highSeverityVulns))
			if cfg.Verification.StrictMode {
				return fmt.Errorf("%d high/critical vulnerabilities found (blocked in strict mode)", highSeverityVulns)
			}
		}
	}

	return nil
}
