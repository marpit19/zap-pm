package main

import (
	"fmt"
	"os"

	"github.com/marpit19/zap-pm/internal/cli"
	"github.com/marpit19/zap-pm/internal/logger"
)


func main() {
	// Initialize logger
	log := logger.New()

	// Initialize and execute root command
	rootCmd := cli.NewRootCommand(log)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
