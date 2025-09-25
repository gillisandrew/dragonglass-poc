package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"dagger.io/dagger"
	"dagger.io/dagger/dag"
	"github.com/spf13/cobra"
)

var (
	ref       string
	commit    string
	directory string
	outputDir string
	buildDir  string

	// Build-time variables (injected via -ldflags)
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "dragonglass-build <path>",
		Short: "Build plugins using Dagger",
		Long:  "A CLI tool to build plugins from a local directory or remote git repository using Dagger",
		Args:  cobra.ExactArgs(1),
		Example: `  # Build from remote repository
  dragonglass-build https://github.com/user/repo.git --ref main --directory plugin-folder
  dragonglass-build https://github.com/user/repo.git --ref main  # uses repository root
  dragonglass-build https://github.com/user/repo.git --commit abc123def456  # build from specific commit
  dragonglass-build https://github.com/user/repo.git --commit abc123def456 --directory plugin-folder
  
  # Build from local directory
  dragonglass-build . --directory example-plugin  # build from ./example-plugin subdirectory
  dragonglass-build /path/to/project --directory my-plugin  # build from /path/to/project/my-plugin
  dragonglass-build ./example-plugin  # build from ./example-plugin (no subdirectory)`,
		Run: func(cmd *cobra.Command, args []string) {
			path := args[0]

			// Validate that both --ref and --commit are not used together
			if ref != "main" && commit != "" {
				fmt.Fprintf(os.Stderr, "Warning: Both --ref and --commit specified. Using commit hash (%s) and ignoring ref (%s).\n", commit, ref)
			}

			// Use directory flag for both local and remote (defaults to root)
			finalDirectory := directory
			if finalDirectory == "" {
				finalDirectory = "." // Use root of the path
			}

			if err := build(context.Background(), path, ref, commit, finalDirectory, outputDir, buildDir); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	// Version command
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("dragonglass-build version %s\n", Version)
			fmt.Printf("Git commit: %s\n", Commit)
			fmt.Printf("Build time: %s\n", BuildTime)
		},
	}

	rootCmd.AddCommand(versionCmd)

	rootCmd.Flags().StringVarP(&ref, "ref", "r", "main", "Git reference (branch or tag) - only used for remote repositories")
	rootCmd.Flags().StringVarP(&commit, "commit", "c", "", "Specific commit hash to use - only used for remote repositories (takes precedence over --ref)")
	rootCmd.Flags().StringVarP(&directory, "directory", "d", "", "Subdirectory to build from (defaults to root of path for both local and remote)")
	rootCmd.Flags().StringVarP(&outputDir, "output-dir", "o", "dist", "Directory where final built plugin artifacts will be exported")
	rootCmd.Flags().StringVar(&buildDir, "build-dir", "", "Directory where npm run build outputs artifacts (relative to plugin directory)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func build(ctx context.Context, path, ref, commit, directory, outputDir, buildDir string) error {
	fmt.Println("Building with Dagger")
	defer dag.Close()

	// create empty directory to put build outputs
	outputs := dag.Directory()

	var workingDir *dagger.Directory

	// Determine if path is a remote repository URL or local directory
	if isRemoteRepository(path) {
		// Use commit if provided, otherwise use ref
		gitRef := ref
		if commit != "" {
			gitRef = commit
			fmt.Printf("Building from remote repository: %s (commit: %s)\n", path, commit)
		} else {
			fmt.Printf("Building from remote repository: %s (ref: %s)\n", path, ref)
		}

		repo := dag.Git(path).Ref(gitRef).Tree()

		if directory == "." {
			fmt.Printf("Using repository root\n")
			workingDir = repo
		} else {
			fmt.Printf("Using directory: %s\n", directory)
			workingDir = repo.Directory(directory)
		}
	} else {
		fmt.Printf("Building from local directory: %s\n", path)
		// Convert relative path to absolute path
		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("failed to resolve absolute path: %v", err)
		}

		// Check if the directory exists
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			return fmt.Errorf("directory does not exist: %s", absPath)
		}

		// For local builds, use the directory flag to specify subdirectory
		repo := dag.Host().Directory(absPath)
		if directory == "." {
			fmt.Printf("Using entire directory\n")
			workingDir = repo
		} else {
			fmt.Printf("Using subdirectory: %s\n", directory)
			workingDir = repo.Directory(directory)
		}
	}

	installer := dag.Container().
		From("node:22").
		WithDirectory("/usr/src/plugin", workingDir).
		WithWorkdir("/usr/src/plugin").
		WithExec([]string{"bash", "-c", "test -f package-lock.json && npm ci || npm install"}).
		WithExec([]string{"bash", "-c", "npm sbom --sbom-type application --sbom-format spdx > sbom.spdx.json"})
		// With([]string{""npm", "sbom", "--sbom-type", "application", "--sbom-format", "spdx", ">", "sbom.spdx.json"}).Terminal()

	builder := installer.WithEnvVariable("NODE_ENV", "production").
		WithExec([]string{"npm", "run", "build"})

	outputs = outputs.WithFile("main.js", builder.File(filepath.Join(buildDir, "main.js"))).
		WithFile("styles.css", builder.File(filepath.Join(buildDir, "styles.css"))).
		WithFile("manifest.json", builder.File("manifest.json")).
		WithFile("sbom.spdx.json", installer.File("sbom.spdx.json"))

	_, err := outputs.Export(ctx, outputDir)
	if err != nil {
		return err
	}
	return nil
}

// isRemoteRepository checks if the given path is a remote repository URL
func isRemoteRepository(path string) bool {
	return strings.HasPrefix(path, "http://") ||
		strings.HasPrefix(path, "https://") ||
		strings.HasPrefix(path, "git@") ||
		strings.HasPrefix(path, "ssh://")
}
