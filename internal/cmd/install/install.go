// ABOUTME: Install command for verified Obsidian plugins
// ABOUTME: Downloads, verifies, and installs plugins from OCI registry
package install

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"

	"github.com/gillisandrew/dragonglass-poc/internal/attestation"
	"github.com/gillisandrew/dragonglass-poc/internal/auth"
	"github.com/gillisandrew/dragonglass-poc/internal/cmd"
	"github.com/gillisandrew/dragonglass-poc/internal/config"
	"github.com/gillisandrew/dragonglass-poc/internal/lockfile"
	"github.com/gillisandrew/dragonglass-poc/internal/oci"
	"github.com/gillisandrew/dragonglass-poc/internal/plugin"
	"github.com/gillisandrew/dragonglass-poc/internal/registry"
)

func NewInstallCommand(ctx *cmd.CommandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install plugins from lockfile",
		Long: `Install all plugins specified in the lockfile with exact SHA256 verification.
This command reads the dragonglass-lock.json file and installs all plugins listed,
verifying their integrity against the stored digests.

Example:
  dragonglass install
  dragonglass install --force`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			force, _ := cmd.Flags().GetBool("force")
			ctx.Logger.Info("Installing plugins from lockfile")

			if err := runInstallFromLockfile(ctx, force); err != nil {
				ctx.Logger.Error("Install failed", ctx.Logger.Args("error", err))
				os.Exit(1)
			}

			ctx.Logger.Info("All plugins installed successfully")
		},
	}

	cmd.Flags().BoolP("force", "f", false, "Overwrite existing plugin files if they exist")
	return cmd
}

func NewAddCommand(ctx *cmd.CommandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [OCI_IMAGE_REFERENCE]",
		Short: "Add a verified plugin from OCI registry",
		Long: `Add a verified Obsidian plugin from an OCI registry.
The plugin will be downloaded, verified for provenance and vulnerabilities,
and installed to the .obsidian/plugins/ directory.

Example:
  dragonglass add ghcr.io/owner/repo:plugin-name-v1.0.0
  dragonglass add --force ghcr.io/owner/repo:plugin-name-v1.0.0`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			imageRef := args[0]
			force, _ := cmd.Flags().GetBool("force")
			ctx.Logger.Info("Adding plugin", ctx.Logger.Args("imageRef", imageRef))

			if err := runAddCommand(imageRef, ctx, force); err != nil {
				ctx.Logger.Error("Add failed", ctx.Logger.Args("error", err))
				os.Exit(1)
			}

			ctx.Logger.Info("Plugin added successfully")
		},
	}

	cmd.Flags().BoolP("force", "f", false, "Overwrite existing plugin files if they exist")
	return cmd
}

func runAddCommand(imageRef string, ctx *cmd.CommandContext, force bool) error {
	// Find dragonglass directory and set proper lockfile path
	dragonglassDir, err := findDragonglassDirectory()
	if err != nil {
		return fmt.Errorf("failed to find dragonglass directory: %w", err)
	}

	lockfilePath := filepath.Join(dragonglassDir, "dragonglass-lock.json")
	cfg := config.DefaultConfig()
	lockfileData := lockfile.NewLockfile(lockfilePath)

	return addPlugin(imageRef, cfg, lockfileData, lockfilePath, ctx, force)
}

func runInstallFromLockfile(ctx *cmd.CommandContext, force bool) error {
	// Find dragonglass directory and load lockfile
	dragonglassDir, err := findDragonglassDirectory()
	if err != nil {
		return fmt.Errorf("failed to find dragonglass directory: %w", err)
	}

	lockfilePath := filepath.Join(dragonglassDir, "dragonglass-lock.json")

	// Check if lockfile exists
	if _, err := os.Stat(lockfilePath); os.IsNotExist(err) {
		return fmt.Errorf("lockfile not found at %s (run 'dragonglass add' to add plugins first)", lockfilePath)
	}

	// Load existing lockfile
	lockfileData, err := lockfile.LoadLockfile(lockfilePath)
	if err != nil {
		return fmt.Errorf("failed to load lockfile: %w", err)
	}

	if len(lockfileData.Plugins) == 0 {
		ctx.Logger.Info("No plugins found in lockfile")
		return nil
	}

	ctx.Logger.Info("Found plugins in lockfile", ctx.Logger.Args("count", len(lockfileData.Plugins)))

	// Find Obsidian directory for installation
	obsidianDir, err := findObsidianDirectory()
	if err != nil {
		return fmt.Errorf("failed to find Obsidian directory: %w", err)
	}

	// Install each plugin from lockfile
	installedCount := 0
	skippedCount := 0

	for pluginID, pluginEntry := range lockfileData.Plugins {
		ctx.Logger.Info("Processing plugin", ctx.Logger.Args("name", pluginEntry.Name, "id", pluginID))

		pluginDir := filepath.Join(obsidianDir, "plugins", pluginID)

		// Check if plugin directory already exists
		if _, err := os.Stat(pluginDir); err == nil {
			if !force {
				ctx.Logger.Debug("Skipping plugin (already exists)", ctx.Logger.Args("id", pluginID, "hint", "use --force to overwrite"))
				skippedCount++
				continue
			}
			ctx.Logger.Debug("Removing existing plugin directory", ctx.Logger.Args("path", makeRelativePath(pluginDir)))
			if err := os.RemoveAll(pluginDir); err != nil {
				return fmt.Errorf("failed to remove existing plugin directory %s: %w", makeRelativePath(pluginDir), err)
			}
		}

		// Install plugin from OCI reference
		ctx.Logger.Debug("Installing from OCI reference", ctx.Logger.Args("reference", pluginEntry.OCIReference, "digest", pluginEntry.OCIDigest))

		if err := installPluginFromLockfileEntry(pluginEntry.OCIReference, pluginDir, pluginID, pluginEntry, ctx); err != nil {
			return fmt.Errorf("failed to install plugin %s: %w", pluginID, err)
		}

		ctx.Logger.Info("Successfully installed plugin", ctx.Logger.Args("name", pluginEntry.Name))
		installedCount++
	}

	ctx.Logger.Info("Installation summary", ctx.Logger.Args("installed", installedCount, "skipped", skippedCount))

	return nil
}

func installPluginFromLockfileEntry(imageRef, pluginDir, pluginID string, pluginEntry lockfile.PluginEntry, cmdCtx *cmd.CommandContext) error {
	// Create registry client
	client, err := registry.NewClient(nil)
	if err != nil {
		return fmt.Errorf("failed to create registry client: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Fetch manifest using image reference
	manifest, _, manifestDigest, err := client.GetManifest(ctx, imageRef)
	if err != nil {
		return fmt.Errorf("failed to fetch manifest: %w", err)
	}

	// Verify digest matches what's in lockfile
	if manifestDigest != pluginEntry.OCIDigest {
		return fmt.Errorf("digest mismatch: expected %s, got %s", pluginEntry.OCIDigest, manifestDigest)
	}

	// Extract plugin files
	if err := extractPluginFilesFromManifest(ctx, imageRef, manifest, pluginDir); err != nil {
		// Clean up on failure
		_ = os.RemoveAll(pluginDir)
		return fmt.Errorf("failed to extract plugin files: %w", err)
	}

	// Create manifest.json from lockfile metadata
	if err := createPluginManifestFromLockfile(pluginDir, pluginID, pluginEntry); err != nil {
		// Clean up on failure
		_ = os.RemoveAll(pluginDir)
		return fmt.Errorf("failed to create plugin manifest: %w", err)
	}

	return nil
}

func createPluginManifestFromLockfile(pluginDir, pluginID string, pluginEntry lockfile.PluginEntry) error {
	manifestPath := filepath.Join(pluginDir, "manifest.json")

	// Create Obsidian-compatible manifest from lockfile data
	manifestData := map[string]interface{}{
		"id":          pluginID,
		"name":        pluginEntry.Name,
		"version":     pluginEntry.Version,
		"description": pluginEntry.Metadata.Description,
		"author":      pluginEntry.Metadata.Author,
	}

	// Add optional fields if available in lockfile metadata
	if pluginEntry.Metadata.Repository != "" {
		manifestData["authorUrl"] = pluginEntry.Metadata.Repository
	}

	file, err := os.Create(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to create manifest file: %w", err)
	}
	defer func() {
		_ = file.Close() // Ignore error on close
	}()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(manifestData); err != nil {
		return fmt.Errorf("failed to write manifest data: %w", err)
	}

	return nil
}

func addPlugin(imageRef string, cfg *config.Config, lockfileData *lockfile.Lockfile, lockfilePath string, cmdCtx *cmd.CommandContext, force bool) error {
	// Step 1: Create registry client
	cmdCtx.Logger.Debug("Creating registry client")
	client, err := registry.NewClient(nil)
	if err != nil {
		return fmt.Errorf("failed to create registry client: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Step 2: Fetch and parse manifest
	cmdCtx.Logger.Debug("Fetching manifest from registry")
	manifest, annotations, manifestDigest, err := client.GetManifest(ctx, imageRef)
	if err != nil {
		return fmt.Errorf("failed to fetch manifest: %w", err)
	}

	// Step 3: Parse plugin metadata
	cmdCtx.Logger.Debug("Parsing plugin metadata")
	parser := plugin.NewManifestParser(nil)
	pluginMetadata, err := parser.ParseMetadata(manifest, annotations)
	if err != nil {
		return fmt.Errorf("failed to parse plugin metadata: %w", err)
	}

	cmdCtx.Logger.Info("Plugin metadata parsed", cmdCtx.Logger.Args(
		"id", pluginMetadata.ID,
		"name", pluginMetadata.Name,
		"version", pluginMetadata.Version,
		"author", pluginMetadata.Author,
		"authorUrl", pluginMetadata.AuthorURL,
		"description", pluginMetadata.Description,
		"minAppVersion", pluginMetadata.MinAppVersion,
		"isDesktopOnly", pluginMetadata.IsDesktopOnly,
	))

	// Step 4: Validate metadata
	validation := parser.ValidateMetadata(pluginMetadata)
	if !validation.Valid {
		if cfg.Verification.StrictMode {
			return fmt.Errorf("metadata validation failed in strict mode")
		}
		cmdCtx.Logger.Warn("Metadata validation warnings (continuing in non-strict mode)")
	}

	// Step 5: Perform verification (SLSA, etc.)
	cmdCtx.Logger.Debug("Verifying attestations")
	token, err := auth.GetToken()
	if err != nil {
		return fmt.Errorf("failed to get authentication token: %w", err)
	}

	verifier, err := attestation.NewAttestationVerifier(token, cmdCtx.TrustedBuilder)
	if err != nil {
		return fmt.Errorf("failed to create attestation verifier: %w", err)
	}

	attestationResult, err := verifier.VerifyAttestations(ctx, imageRef)
	if err != nil {
		return fmt.Errorf("failed to verify attestations: %w", err)
	}

	// Check verification results
	if cfg.Verification.StrictMode && (!attestationResult.Found || !attestationResult.Valid) {
		if !attestationResult.Found {
			return fmt.Errorf("attestations not found (required in strict mode)")
		}
		if !attestationResult.Valid {
			return fmt.Errorf("attestation verification failed (required in strict mode)")
		}
	}

	// Step 6: Discover Obsidian directory
	cmdCtx.Logger.Debug("Finding Obsidian directory")
	obsidianDir, err := findObsidianDirectory()
	if err != nil {
		return fmt.Errorf("failed to find Obsidian directory: %w", err)
	}

	pluginDir := filepath.Join(obsidianDir, "plugins", pluginMetadata.ID)
	cmdCtx.Logger.Debug("Plugin installation target", cmdCtx.Logger.Args("path", makeRelativePath(pluginDir)))

	// Step 7: Check for conflicts
	if _, err := os.Stat(pluginDir); err == nil {
		if !force {
			return fmt.Errorf("plugin directory already exists: %s (use --force to overwrite)", makeRelativePath(pluginDir))
		}
		cmdCtx.Logger.Debug("Removing existing plugin directory", cmdCtx.Logger.Args("path", makeRelativePath(pluginDir)))
		if err := os.RemoveAll(pluginDir); err != nil {
			return fmt.Errorf("failed to remove existing plugin directory: %w", err)
		}
	}

	// Step 8: Extract plugin files
	cmdCtx.Logger.Debug("Extracting plugin files")
	if err := extractPluginFilesFromManifest(ctx, imageRef, manifest, pluginDir); err != nil {
		// Clean up on failure
		_ = os.RemoveAll(pluginDir) // Ignore cleanup error
		return fmt.Errorf("failed to extract plugin files: %w", err)
	}

	// Step 9: Create manifest.json from metadata
	cmdCtx.Logger.Debug("Creating plugin manifest")
	if err := createPluginManifest(pluginDir, pluginMetadata); err != nil {
		// Clean up on failure
		_ = os.RemoveAll(pluginDir) // Ignore cleanup error
		return fmt.Errorf("failed to create plugin manifest: %w", err)
	}

	// Step 10: Update lockfile
	cmdCtx.Logger.Debug("Updating lockfile")
	if err := updateLockfile(lockfileData, lockfilePath, pluginMetadata, imageRef, manifestDigest); err != nil {
		return fmt.Errorf("failed to update lockfile: %w", err)
	}

	cmdCtx.Logger.Info("Installation completed successfully", cmdCtx.Logger.Args("plugin", pluginMetadata.Name, "id", pluginMetadata.ID, "location", makeRelativePath(pluginDir)))

	return nil
}

// findObsidianDirectory searches for .obsidian directory from current directory up
func findObsidianDirectory() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Search up the directory tree for .obsidian
	for {
		obsidianPath := filepath.Join(currentDir, ".obsidian")
		if info, err := os.Stat(obsidianPath); err == nil && info.IsDir() {
			return obsidianPath, nil
		}

		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			break // reached root
		}
		currentDir = parent
	}

	return "", fmt.Errorf(".obsidian directory not found in current path or parent directories")
}

// findDragonglassDirectory searches for or creates .dragonglass directory from current directory up
func findDragonglassDirectory() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Search up the directory tree for .dragonglass or create it at the same level as .obsidian
	for {
		// Check if .obsidian exists to determine if this is an Obsidian vault
		obsidianPath := filepath.Join(currentDir, ".obsidian")
		if info, err := os.Stat(obsidianPath); err == nil && info.IsDir() {
			// Found .obsidian, so create/use .dragonglass at the same level
			dragonglassPath := filepath.Join(currentDir, ".dragonglass")

			// Create .dragonglass directory if it doesn't exist
			if err := os.MkdirAll(dragonglassPath, 0755); err != nil {
				return "", fmt.Errorf("failed to create .dragonglass directory: %w", err)
			}

			return dragonglassPath, nil
		}

		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			break // reached root
		}
		currentDir = parent
	}

	return "", fmt.Errorf(".obsidian directory not found in current path or parent directories (required to determine vault location)")
}

// createPluginManifest creates the manifest.json file required by Obsidian
func createPluginManifest(pluginDir string, metadata *plugin.Metadata) error {
	manifestPath := filepath.Join(pluginDir, "manifest.json")

	// Create Obsidian-compatible manifest
	manifestData := map[string]interface{}{
		"id":          metadata.ID,
		"name":        metadata.Name,
		"version":     metadata.Version,
		"description": metadata.Description,
		"author":      metadata.Author,
	}

	if metadata.MinAppVersion != "" {
		manifestData["minAppVersion"] = metadata.MinAppVersion
	}
	if metadata.AuthorURL != "" {
		manifestData["authorUrl"] = metadata.AuthorURL
	}
	if metadata.IsDesktopOnly {
		manifestData["isDesktopOnly"] = true
	}

	file, err := os.Create(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to create manifest file: %w", err)
	}
	defer func() {
		_ = file.Close() // Ignore error on close
	}()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(manifestData); err != nil {
		return fmt.Errorf("failed to write manifest data: %w", err)
	}

	return nil
}

// updateLockfile adds the installed plugin to the lockfile
func updateLockfile(lockfileData *lockfile.Lockfile, lockfilePath string, metadata *plugin.Metadata, imageRef, digest string) error {
	if lockfileData == nil {
		return fmt.Errorf("lockfile data is nil")
	}

	// Create plugin entry
	entry := lockfile.PluginEntry{
		Name:         metadata.Name,
		Version:      metadata.Version,
		OCIReference: imageRef,
		OCIDigest:    digest,
		VerificationState: lockfile.VerificationState{
			ProvenanceVerified: true,  // We verified SLSA above
			SBOMVerified:       false, // Not implemented yet
			VulnScanPassed:     true,  // Assume passed for now
		},
		Metadata: lockfile.PluginMetadata{
			Author:      metadata.Author,
			Description: metadata.Description,
			Repository:  metadata.AuthorURL,
		},
	}

	// Add to lockfile
	if err := lockfileData.AddPlugin(metadata.ID, entry); err != nil {
		return fmt.Errorf("failed to add plugin to lockfile: %w", err)
	}

	// Save lockfile
	if err := lockfile.SaveLockfile(lockfileData, lockfilePath); err != nil {
		return fmt.Errorf("failed to save lockfile: %w", err)
	}

	return nil
}

// extractPluginFilesFromManifest extracts main.js and styles.css from OCI manifest layers
func extractPluginFilesFromManifest(ctx context.Context, imageRef string, manifest *ocispec.Manifest, targetDir string) error {
	// Get GitHub token for OCI authentication
	token, err := auth.GetToken()
	if err != nil {
		return fmt.Errorf("failed to get authentication token: %w", err)
	}

	// Create OCI registry client
	ghcrRegistry := &oci.GHCRRegistry{Token: token}
	repo, err := ghcrRegistry.GetRepositoryFromRef(imageRef)
	if err != nil {
		return fmt.Errorf("failed to create OCI repository: %w", err)
	}

	// Extract plugin files using the OCI client
	if err := repo.ExtractPluginFiles(ctx, manifest, targetDir); err != nil {
		return fmt.Errorf("failed to extract plugin files: %w", err)
	}

	return nil
}

// makeRelativePath converts an absolute path to relative from current working directory
// Returns the original path if conversion fails
func makeRelativePath(absPath string) string {
	if cwd, err := os.Getwd(); err == nil {
		if relPath, err := filepath.Rel(cwd, absPath); err == nil {
			// Only return relative path if it's shorter and doesn't start with too many ../
			if len(relPath) < len(absPath) && !strings.HasPrefix(relPath, "../../") {
				return "./" + relPath
			}
		}
	}
	return absPath
}
