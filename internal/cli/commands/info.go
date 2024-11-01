package commands

import (
	"fmt"

	"github.com/marpit19/zap-pm/internal/logger"
	"github.com/marpit19/zap-pm/internal/registry"
	"github.com/spf13/cobra"
)

// NewInfoCmd creates a new info command
func NewInfoCmd(log *logger.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info [package]",
		Short: "Display information about a package",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			packageName := args[0]
			registryClient := registry.NewRegistryClient(log)

			// Fetch package metadata
			metadata, err := registryClient.GetPackageMetadata(packageName)
			if err != nil {
				return fmt.Errorf("failed to fetch package info: %w", err)
			}

			// Display package information
			fmt.Printf("Package: %s\n", metadata.Name)
			fmt.Printf("Latest Version: %s\n", metadata.DistTags["latest"])

			// Display versions
			fmt.Println("\nAvailable Versions:")
			for version := range metadata.Versions {
				fmt.Printf("- %s\n", version)
			}

			// Get latest version info
			latest, err := registryClient.GetLatestVersion(packageName)
			if err != nil {
				return fmt.Errorf("failed to fetch latest version info: %w", err)
			}

			// Display dependencies
			if len(latest.Dependencies) > 0 {
				fmt.Printf("\nDependencies (latest version):\n")
				for dep, ver := range latest.Dependencies {
					fmt.Printf("- %s: %s\n", dep, ver)
				}
			}

			return nil
		},
	}

	return cmd
}
