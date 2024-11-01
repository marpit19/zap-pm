package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/marpit19/zap-pm/internal/errors"
)

// ValidationError represents a package.json validation error
type ValidationError struct {
	Field   string
	Message string
}

// Validate performs validation on PackageJSON fields
func (p *PackageJSON) Validate() []ValidationError {
	var validationErrors []ValidationError

	// Validate name
	if err := validateName(p.Name); err != nil {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "name",
			Message: err.Error(),
		})
	}

	// Validate version
	if err := validateVersion(p.Version); err != nil {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "version",
			Message: err.Error(),
		})
	}

	// Validate dependencies
	for dep, ver := range p.Dependencies {
		if err := validateDependency(dep, ver); err != nil {
			validationErrors = append(validationErrors, ValidationError{
				Field:   "dependencies." + dep,
				Message: err.Error(),
			})
		}
	}

	return validationErrors
}

// validateName checks if the package name is valid
func validateName(name string) error {
	if name == "" {
		return errors.New(errors.ErrInvalidPackageJSON, "name cannot be empty", nil)
	}

	if len(name) > 214 {
		return errors.New(errors.ErrInvalidPackageJSON, "name too long (max 214 chars)", nil)
	}

	// Name must match npm package name rules
	valid := regexp.MustCompile(`^(?:@[a-z0-9-*~][a-z0-9-*._~]*\/)?[a-z0-9-~][a-z0-9-._~]*$`)
	if !valid.MatchString(name) {
		return errors.New(errors.ErrInvalidPackageJSON, "invalid package name format", nil)
	}

	return nil
}

// validateVersion checks if the version string is valid
func validateVersion(version string) error {
	if version == "" {
		return errors.New(errors.ErrInvalidPackageJSON, "version cannot be empty", nil)
	}

	// Basic semver format check
	valid := regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)
	if !valid.MatchString(version) {
		return errors.New(errors.ErrInvalidPackageJSON, "invalid version format (must be semver)", nil)
	}

	return nil
}

// validateDependency checks if the dependency version string is valid
func validateDependency(name, version string) error {
	if strings.TrimSpace(version) == "" {
		return errors.New(errors.ErrInvalidPackageJSON,
			fmt.Sprintf("version for dependency '%s' cannot be empty", name),
			nil)
	}

	valid := regexp.MustCompile(`^([\^~]?[0-9]+\.[0-9]+\.[0-9]+|latest|[*]|>=[0-9]+\.[0-9]+\.[0-9]+|file:.*|git\+https:\/\/.*)$`)
	if !valid.MatchString(version) {
		return errors.New(errors.ErrInvalidPackageJSON,
			fmt.Sprintf("invalid version format for dependency '%s'", name),
			nil)
	}
	return nil
}
