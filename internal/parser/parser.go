package parser

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/marpit19/zap-pm/internal/errors"
)

// PackageJSON represents the structure of a package.json file
type PackageJSON struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Description     string            `json:"description,omitempty"`
	Main            string            `json:"main,omitempty"`
	Scripts         map[string]string `json:"scripts,omitempty"`
	Dependencies    map[string]string `json:"dependencies,omitempty"`
	DevDependencies map[string]string `json:"devDependencies,omitempty"`
}

// DefaultPackageJSON creates a package.json with default values
func DefaultPackageJSON() *PackageJSON {
	return &PackageJSON{
		Name:            filepath.Base(getCurrentDir()),
		Version:         "1.0.0",
		Description:     "",
		Main:            "index.js",
		Scripts:         make(map[string]string),
		Dependencies:    make(map[string]string),
		DevDependencies: make(map[string]string),
	}
}

// ParsePackageJSON reads and parses a package.json file
func ParsePackageJSON(filename string) (*PackageJSON, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, errors.New(errors.ErrPackageJSONNotFound, "failed to read package.json", err)
	}

	var pkg PackageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, errors.New(errors.ErrInvalidPackageJSON, "failed to parse package.json", err)
	}

	return &pkg, nil
}

// WriteToFile writes the PackageJSON to a file
func (p *PackageJSON) WriteToFile(filename string) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return errors.New(errors.ErrInvalidPackageJSON, "failed to marshal package.json", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return errors.New(errors.ErrInvalidPackageJSON, "failed to write package.json", err)
	}

	return nil
}

// getCurrentDir returns the current directory name
func getCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "my-project"
	}
	return dir
}
