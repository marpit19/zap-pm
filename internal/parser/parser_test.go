package parser

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteAndParsePackageJSON(t *testing.T) {
	// Create a temporary package.json
	pkg := PackageJSON{
		Name:    "test-package",
		Version: "1.0.0",
		Dependencies: map[string]string{
			"express": "^4.17.1",
		},
	}

	tmpfile := "test_package.json"

	// Write the package.json
	err := pkg.WriteToFile(tmpfile)
	assert.NoError(t, err)

	// Read it back
	parsedPkg, err := ParsePackageJSON(tmpfile)
	assert.NoError(t, err)

	// Compare
	assert.Equal(t, pkg.Name, parsedPkg.Name)
	assert.Equal(t, pkg.Version, parsedPkg.Version)
	assert.Equal(t, pkg.Dependencies["express"], parsedPkg.Dependencies["express"])

	// Cleanup
	os.Remove(tmpfile)
}

func TestInvalidPackageJSON(t *testing.T) {
	testCases := []struct {
		name     string
		pkg      PackageJSON
		expected string
	}{
		{
			name: "invalid name",
			pkg: PackageJSON{
				Name:    "Invalid Name", // contains space
				Version: "1.0.0",
			},
			expected: "invalid package name format",
		},
		{
			name: "invalid version",
			pkg: PackageJSON{
				Name:    "valid-name",
				Version: "1.0", // incomplete version
			},
			expected: "invalid version format",
		},
		{
			name: "empty dependency version",
			pkg: PackageJSON{
				Name:    "valid-name",
				Version: "1.0.0",
				Dependencies: map[string]string{
					"express": "",
				},
			},
			expected: "cannot be empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.pkg.WriteToFile("test_package.json")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expected)
		})
	}
}

func TestDefaultPackageJSON(t *testing.T) {
	pkg := DefaultPackageJSON()
	assert.NotEmpty(t, pkg.Name)
	assert.Equal(t, "1.0.0", pkg.Version)
	assert.NotNil(t, pkg.Dependencies)
	assert.NotNil(t, pkg.DevDependencies)
}
