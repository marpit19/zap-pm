package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/marpit19/zap-pm/internal/downloader"
	"github.com/marpit19/zap-pm/internal/logger"
	"github.com/marpit19/zap-pm/internal/registry"
	"github.com/spf13/cobra"
)

// NewDownloadCmd creates a new download command
func NewDownloadCmd(log *logger.Logger) *cobra.Command {
	var withDependencies bool

	cmd := &cobra.Command{
		Use:   "download [package[@version]]",
		Short: "Download a package",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse package name and version
			packageName, version := parsePackageArg(args[0])

			// Set up registry client
			registryClient := registry.NewRegistryClient(log)

			// If no version specified, get latest
			if version == "" {
				latestInfo, err := registryClient.GetLatestVersion(packageName)
				if err != nil {
					return fmt.Errorf("failed to get latest version: %w", err)
				}
				version = latestInfo.Version
			}

			// Create cache directory
			cacheDir := getCacheDir()
			if err := os.MkdirAll(cacheDir, 0755); err != nil {
				return fmt.Errorf("failed to create cache directory: %w", err)
			}

			// Create download manager
			dm := downloader.NewDownloadManager(registryClient, cacheDir, log)

			// Set download options
			opts := downloader.DownloadOptions{
				UseCache:     true,
				ShowProgress: true,
				Concurrency:  3,
			}

			// Download package
			if withDependencies {
				log.Infof("Downloading %s@%s with dependencies...", packageName, version)
				results, err := dm.DownloadDependencies(packageName, version, opts)
				if err != nil {
					return fmt.Errorf("failed to download dependencies: %w", err)
				}
				log.Infof("Successfully downloaded %d packages", len(results))
			} else {
				log.Infof("Downloading %s@%s...", packageName, version)
				result, err := dm.DownloadPackage(packageName, version, opts)
				if err != nil {
					return fmt.Errorf("failed to download package: %w", err)
				}
				log.Infof("Successfully downloaded to: %s", result.Path)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&withDependencies, "with-dependencies", false, "Download package dependencies")
	return cmd
}

// NewVerifyCmd creates a new verify command
func NewVerifyCmd(log *logger.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify [package[@version]]",
		Short: "Verify package integrity in cache",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			packageName, version := parsePackageArg(args[0])

			registryClient := registry.NewRegistryClient(log)
			dm := downloader.NewDownloadManager(registryClient, getCacheDir(), log)

			// Check if package exists in cache
			opts := downloader.DownloadOptions{UseCache: true}
			result, err := dm.DownloadPackage(packageName, version, opts)
			if err != nil {
				return fmt.Errorf("failed to verify package: %w", err)
			}

			log.Infof("Package %s@%s verified successfully", packageName, version)
			log.Infof("Location: %s", result.Path)
			log.Infof("Checksum: %s", result.Shasum)

			return nil
		},
	}

	return cmd
}

// Helper functions

func parsePackageArg(arg string) (name, version string) {
	parts := strings.Split(arg, "@")
	if len(parts) > 1 {
		return parts[0], parts[1]
	}
	return parts[0], ""
}

func getCacheDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	return filepath.Join(homeDir, ".zap", "cache")
}
