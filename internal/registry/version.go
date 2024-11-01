package registry

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// resolveVersion resolves a version constraint to a specific version
func (c *RegistryClient) resolveVersion(name, versionConstraint string) (string, error) {
	c.log.Debugf("Resolving version constraint %s for package %s", versionConstraint, name)

	// If it's an exact version, just return it
	if isExactVersion(versionConstraint) {
		return versionConstraint, nil
	}

	// Get package metadata
	metadata, err := c.GetPackageMetadata(name)
	if err != nil {
		return "", fmt.Errorf("failed to get package metadata: %w", err)
	}

	// Parse version constraint
	constraint, err := semver.NewConstraint(versionConstraint)
	if err != nil {
		return "", fmt.Errorf("invalid version constraint %s: %w", versionConstraint, err)
	}

	// Get all available versions
	var versions []*semver.Version
	for version := range metadata.Versions {
		v, err := semver.NewVersion(version)
		if err != nil {
			c.log.Debugf("Skipping invalid version %s: %v", version, err)
			continue
		}
		versions = append(versions, v)
	}

	// Sort versions in descending order (highest first)
	sort.Sort(sort.Reverse(semver.Collection(versions)))

	// Find the highest version that satisfies the constraint
	for _, v := range versions {
		if constraint.Check(v) {
			resolvedVersion := v.String()
			c.log.Debugf("Resolved %s to version %s", versionConstraint, resolvedVersion)
			return resolvedVersion, nil
		}
	}

	return "", fmt.Errorf("no version found matching constraint %s", versionConstraint)
}

// isExactVersion checks if the version string is an exact version
func isExactVersion(version string) bool {
	// Remove any whitespace
	version = strings.TrimSpace(version)

	// Check if it starts with any range indicators
	if strings.HasPrefix(version, "^") ||
		strings.HasPrefix(version, "~") ||
		strings.HasPrefix(version, ">") ||
		strings.HasPrefix(version, "<") ||
		strings.HasPrefix(version, "=") ||
		version == "latest" ||
		version == "*" {
		return false
	}

	// Try to parse as semver
	_, err := semver.NewVersion(version)
	return err == nil
}
