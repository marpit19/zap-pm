package commands

import (
	"os"

	"github.com/marpit19/zap-pm/internal/errors"
	"github.com/marpit19/zap-pm/internal/logger"
	"github.com/marpit19/zap-pm/internal/parser"
	"github.com/spf13/cobra"
)

// NewInitCmd creates a new init command
func NewInitCmd(log *logger.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a new package.json",
		Long:  `Creates a new package.json file in the current directory if one doesn't exist.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if package.json already exists
			if _, err := os.Stat("package.json"); err == nil {
				log.Info("package.json already exists")
				return nil
			}

			// Create default package.json
			pkg := parser.DefaultPackageJSON()

			// Write to file
			if err := pkg.WriteToFile("package.json"); err != nil {
				return errors.Wrap(err, "failed to create package.json")
			}

			log.Info("Created package.json")
			return nil
		},
	}
}
