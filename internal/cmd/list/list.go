// ABOUTME: List command for displaying installed verified plugins
// ABOUTME: Shows plugins from lockfile with their verification status
package list

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/gillisandrew/dragonglass-poc/internal/cmd"
	"github.com/gillisandrew/dragonglass-poc/internal/config"
	"github.com/gillisandrew/dragonglass-poc/internal/lockfile"
)

func NewListCommand(ctx *cmd.CommandContext) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed verified plugins",
		Long: `List all plugins installed through dragonglass in the current vault.
Displays plugin names, versions, installation status, and verification details
from the lockfile.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := runListCommand(ctx); err != nil {
				ctx.Logger.Error("List command failed", ctx.Logger.Args("error", err))
				os.Exit(1)
			}
		},
	}
}

func runListCommand(ctx *cmd.CommandContext) error {
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

	// Find dragonglass directory and load lockfile (same logic as install/add commands)
	dragonglassDir, err := findDragonglassDirectory()
	if err != nil {
		return fmt.Errorf("failed to find dragonglass directory: %w", err)
	}

	lockfilePath := filepath.Join(dragonglassDir, "dragonglass-lock.json")

	// Check if lockfile exists
	if _, err := os.Stat(lockfilePath); os.IsNotExist(err) {
		return fmt.Errorf("no lockfile found at %s (run 'dragonglass add' to add plugins first)", lockfilePath)
	}

	// Load existing lockfile
	lockfileData, err := lockfile.LoadLockfile(lockfilePath)
	if err != nil {
		return fmt.Errorf("failed to load lockfile: %w", err)
	}

	if len(lockfileData.Plugins) == 0 {
		ctx.Logger.Info("No verified plugins installed in this vault")
		return nil
	}

	if cfg.Output.Format == "json" {
		ctx.Logger.Warn("JSON output not yet implemented", ctx.Logger.Args("pluginCount", len(lockfileData.Plugins)))
		return nil
	}

	// Build table data
	tableData := pterm.TableData{
		{"ID", "NAME", "VERSION", "VERIFIED", "STATUS", "OCI REFERENCE"},
	}

	for pluginID, plugin := range lockfileData.Plugins {
		status := "OK"
		if len(plugin.VerificationState.Errors) > 0 {
			status = "ERROR"
		} else if len(plugin.VerificationState.Warnings) > 0 {
			status = "WARNING"
		}

		verifiedStatus := "No"
		if plugin.VerificationState.ProvenanceVerified && plugin.VerificationState.SBOMVerified {
			verifiedStatus = "Yes"
		}

		tableData = append(tableData, []string{
			pluginID,
			plugin.Name,
			plugin.Version,
			verifiedStatus,
			status,
			plugin.OCIReference,
		})
	}

	// Render table with pterm
	pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()

	ctx.Logger.Info("Plugin list summary", ctx.Logger.Args("total", len(lockfileData.Plugins), "lockfile", lockfilePath))

	return nil
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
