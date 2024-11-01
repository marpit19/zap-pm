package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/marpit19/zap-pm/internal/errors"
	"github.com/marpit19/zap-pm/internal/logger"
	"github.com/marpit19/zap-pm/internal/parser"
	"github.com/spf13/cobra"
)

// promptUser asks for user confirmation
func promptUser(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [y/N]: ", prompt)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// NewInitCmd creates a new init command
func NewInitCmd(log *logger.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a new package.json",
		Long:  `Creates a new package.json file in the current directory if one doesn't exist, or validates existing package.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := os.Stat("package.json"); err == nil {
				// Try to parse existing package.json
				existingPkg, err := parser.ParsePackageJSON("package.json")
				if err != nil {
					log.Error("Invalid package.json detected")
					fmt.Println("\nValidation errors:")
					if valErrs, ok := err.(*errors.ZapError); ok {
						fmt.Printf("- %s\n", valErrs.Message)
					}

					if !promptUser("Would you like to create a new package.json with default values?") {
						log.Info("Operation cancelled")
						return nil
					}

					// Create new package.json
					pkg := parser.DefaultPackageJSON()
					if err := pkg.WriteToFile("package.json"); err != nil {
						return err
					}
					log.Info("Created new package.json with default values")
					return nil
				}

				// Show existing package details
				log.Info("Existing package.json is valid")
				fmt.Printf("\nCurrent package.json details:\n")
				fmt.Printf("Name: %s\n", existingPkg.Name)
				fmt.Printf("Version: %s\n", existingPkg.Version)
				if len(existingPkg.Dependencies) > 0 {
					fmt.Printf("Dependencies: %d\n", len(existingPkg.Dependencies))
				}
				if len(existingPkg.DevDependencies) > 0 {
					fmt.Printf("DevDependencies: %d\n", len(existingPkg.DevDependencies))
				}

				if promptUser("Would you like to reinitialize package.json with default values?") {
					pkg := parser.DefaultPackageJSON()
					if err := pkg.WriteToFile("package.json"); err != nil {
						return err
					}
					log.Info("Created new package.json with default values")
				} else {
					log.Info("Keeping existing package.json")
				}
				return nil
			}

			// No existing package.json, create new one
			log.Info("No package.json found. Creating new one...")
			pkg := parser.DefaultPackageJSON()

			// Write to file
			if err := pkg.WriteToFile("package.json"); err != nil {
				return err
			}

			log.Info("Created new package.json with default values")
			return nil
		},
	}
}
