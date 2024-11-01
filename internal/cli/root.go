package cli

import (
	"github.com/marpit19/zap-pm/internal/cli/commands"
	"github.com/marpit19/zap-pm/internal/logger"
	"github.com/spf13/cobra"
)

const (
	version = "0.1.0"
)

// NewRootCommand creates the root command
func NewRootCommand(log *logger.Logger) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "zap",
		Short:   "Zap - A fast, simple package manager for JavaScript",
		Version: version,
		// Disable the default completion command
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}

	// Add commands
	rootCmd.AddCommand(
		commands.NewVersionCmd(),
		commands.NewInitCmd(log),
		commands.NewInfoCmd(log),
		commands.NewDownloadCmd(log),
		commands.NewVerifyCmd(log),
	)

	return rootCmd
}
