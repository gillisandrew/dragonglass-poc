// ABOUTME: Shared command context structure for CLI commands
// ABOUTME: Contains global configuration that can be passed to all commands
package cmd

import "github.com/pterm/pterm"

// CommandContext holds global configuration that can be passed to commands
type CommandContext struct {
	AnnotationNamespace string
	TrustedBuilder      string
	ConfigPath          string
	LockfilePath        string
	GitHubToken         string
	Logger              *pterm.Logger
}
