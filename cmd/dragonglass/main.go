// ABOUTME: Main entry point for the dragonglass CLI application
// ABOUTME: Sets up the root command and executes the CLI
package main

import (
	"os"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/gillisandrew/dragonglass-poc/internal/cmd"
	"github.com/gillisandrew/dragonglass-poc/internal/cmd/auth"
	"github.com/gillisandrew/dragonglass-poc/internal/cmd/install"
	"github.com/gillisandrew/dragonglass-poc/internal/cmd/list"
	"github.com/gillisandrew/dragonglass-poc/internal/cmd/verify"
)

var (
	// Global flags
	annotationNamespace string
	trustedBuilder      string
	configPath          string
	lockfilePath        string
	githubToken         string
	verbose             bool
	quiet               bool
)

var rootCmd = &cobra.Command{
	Use:   "dragonglass",
	Short: "A secure Obsidian plugin manager with provenance verification",
	Long: `Dragonglass is a CLI tool that provides secure plugin management for Obsidian
by verifying provenance attestations and Software Bill of Materials (SBOM).

It ensures plugins are built through authorized workflows and performs
vulnerability scanning before installation.`,
}

func init() {
	// Global persistent flags
	rootCmd.PersistentFlags().StringVar(&annotationNamespace, "annotation-namespace", "vnd.obsidian.plugin", "Plugin annotation namespace prefix")
	rootCmd.PersistentFlags().StringVar(&trustedBuilder, "trusted-builder", "https://github.com/gillisandrew/dragonglass-poc/.github/workflows/build.yml@refs/heads/main", "Trusted workflow signer identity")
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "Path to configuration file")
	rootCmd.PersistentFlags().StringVar(&lockfilePath, "lockfile", "", "Path to lockfile")
	rootCmd.PersistentFlags().StringVar(&githubToken, "github-token", "", "GitHub authentication token")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging (debug level)")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Enable quiet mode (warnings and errors only)")
}

// getGitHubToken returns the GitHub token from flag or environment variables
func getGitHubToken(flagToken string) string {
	if flagToken != "" {
		return flagToken
	}

	// Try environment variables as fallback
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token
	}
	if token := os.Getenv("GH_TOKEN"); token != "" {
		return token
	}

	return ""
}

func main() {
	// Initialize logger based on flags
	var logger *pterm.Logger
	if quiet {
		logger = pterm.DefaultLogger.WithTime(false).WithLevel(pterm.LogLevelWarn)
	} else if verbose {
		logger = pterm.DefaultLogger.WithTime(false).WithLevel(pterm.LogLevelDebug)
	} else {
		logger = pterm.DefaultLogger.WithTime(false).WithLevel(pterm.LogLevelInfo)
	}

	// Configure logger to write to stderr to keep stdout clean
	logger = logger.WithWriter(os.Stderr)

	// Initialize command context with global flags
	cmdContext := &cmd.CommandContext{
		AnnotationNamespace: annotationNamespace,
		TrustedBuilder:      trustedBuilder,
		ConfigPath:          configPath,
		LockfilePath:        lockfilePath,
		GitHubToken:         getGitHubToken(githubToken),
		Logger:              logger,
	}

	// Add commands with context
	rootCmd.AddCommand(auth.NewAuthCommand(cmdContext))
	rootCmd.AddCommand(install.NewInstallCommand(cmdContext))
	rootCmd.AddCommand(install.NewAddCommand(cmdContext))
	rootCmd.AddCommand(verify.NewVerifyCommand(cmdContext))
	rootCmd.AddCommand(list.NewListCommand(cmdContext))

	if err := rootCmd.Execute(); err != nil {
		logger.Error("Command execution failed", logger.Args("error", err))
		os.Exit(1)
	}
}
