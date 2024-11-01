package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewVersionCmd creates a new version command
func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number of Zap",
		Long:  `All software has versions. This is Zap's.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Zap Package Manager v%s\n", cmd.Root().Version)
		},
	}
}
