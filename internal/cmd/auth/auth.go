// ABOUTME: Authentication command for GitHub App device flow
// ABOUTME: Handles user authentication to access ghcr.io registry
package auth

import (
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/gillisandrew/dragonglass-poc/internal/cmd"
	"github.com/gillisandrew/dragonglass-poc/internal/config"
)

func NewAuthCommand(ctx *cmd.CommandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with GitHub App using device flow",
		Long: `Authenticate with GitHub App to access ghcr.io registry.
This command will guide you through the OAuth device flow process
and securely store your authentication credentials.

The authentication uses the same proven flow as the GitHub CLI (gh).`,
		Run: func(cmd *cobra.Command, args []string) {
			err := runAuthCommand(ctx)
			if err != nil {
				ctx.Logger.Error("Authentication failed", ctx.Logger.Args("error", err))
				return
			}
		},
	}

	cmd.AddCommand(newStatusCommand(ctx))
	cmd.AddCommand(newLogoutCommand(ctx))

	return cmd
}

func runAuthCommand(ctx *cmd.CommandContext) error {
	authService := ctx.AuthService

	// Check if already authenticated
	if authService.IsAuthenticated() {
		username, err := authService.GetUser()
		if err != nil {
			// Don't fail completely if we can't get username details
			username = "authenticated user"
		}

		// Load configuration to show registry
		configOpts := config.DefaultConfigOpts()
		if ctx.ConfigPath != "" {
			configOpts = configOpts.WithConfigPath(ctx.ConfigPath)
		}
		configManager := config.NewConfigManager(configOpts)
		cfg, _, err := configManager.LoadConfig()
		if err != nil {
			cfg = config.DefaultConfig()
		}

		pterm.Success.Printfln("Already authenticated as %s", pterm.LightCyan(username))
		pterm.Info.Printfln("Registry configured: %s", pterm.LightBlue(cfg.Registry.DefaultRegistry))
		pterm.Info.Println("Use 'dragonglass auth status' to view details")
		return nil
	}

	// Run device flow authentication
	return authService.Authenticate()
}

func newStatusCommand(ctx *cmd.CommandContext) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "View authentication status",
		Long:  `Display current authentication status and user information.`,
		Run: func(cmd *cobra.Command, args []string) {
			authService := ctx.AuthService

			if !authService.IsAuthenticated() {
				pterm.Warning.Printfln("Not authenticated with GitHub")
				pterm.Info.Println("Run 'dragonglass auth' to authenticate")
				return
			}

			username, err := authService.GetUser()
			if err != nil {
				username = "authenticated user"
			}

			token, err := authService.GetToken()
			if err != nil {
				ctx.Logger.Error("Error getting token", ctx.Logger.Args("error", err))
				return
			}

			// Get stored credential details
			cred, err := authService.GetCredential()
			if err != nil {
				ctx.Logger.Error("Error getting credential details", ctx.Logger.Args("error", err))
				return
			}

			// Don't show full token for security
			maskedToken := maskToken(token)

			ctx.Logger.Info("Authentication status",
				ctx.Logger.Args(
					"status", "authenticated",
					"host", cred.Host,
					"username", username,
					"token", maskedToken,
					"scopes", cred.Scopes,
					"registry", "ghcr.io",
					"storage", cred.Source,
					"created", cred.CreatedAt.Format("2006-01-02 15:04:05"),
				),
			)
		},
	}
}

func newLogoutCommand(ctx *cmd.CommandContext) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Sign out and remove stored credentials",
		Long:  `Remove stored authentication credentials and sign out.`,
		Run: func(cmd *cobra.Command, args []string) {
			authService := ctx.AuthService

			if !authService.IsAuthenticated() {
				pterm.Warning.Println("Not currently authenticated")
				return
			}

			// Get username before logout for confirmation
			username, _ := authService.GetUser()
			if username == "" {
				username = "authenticated user"
			}

			// Clear stored credentials
			if err := authService.Logout(); err != nil {
				ctx.Logger.Error("Error clearing credentials", ctx.Logger.Args("error", err))
				return
			}

			pterm.Success.Printfln("Successfully logged out %s", username)
			pterm.Info.Println("All stored credentials have been removed")
		},
	}
}

func maskToken(token string) string {
	if len(token) <= 8 {
		return "********"
	}
	return token[:4] + "..." + token[len(token)-4:]
}
