package commands

import (
	"os"
	"testing"

	"github.com/marpit19/zap-pm/internal/logger"
	"github.com/stretchr/testify/assert"
)

func TestInitCommand(t *testing.T) {
	// Setup
	log := logger.New()
	cmd := NewInitCmd(log)

	// Cleanup any existing package.json
	os.Remove("package.json")

	// Test execution
	err := cmd.Execute()
	assert.NoError(t, err)

	// Verify package.json was created
	_, err = os.Stat("package.json")
	assert.NoError(t, err)

	// Test second execution (should not error when file exists)
	err = cmd.Execute()
	assert.NoError(t, err)

	// Cleanup
	os.Remove("package.json")
}