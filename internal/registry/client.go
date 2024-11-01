package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/marpit19/zap-pm/internal/errors"
	"github.com/marpit19/zap-pm/internal/logger"
)

const (
	defaultRegistryURL = "https://registry.npmjs.org/"
	defaultTimeout     = 30 * time.Second
	defaultRetires     = 3
)

// RetryConfig holds retry-related configuration
type RetryConfig struct {
	MaxRetries  int
	RetryDelay  time.Duration
	MaxWaitTime time.Duration
}

// RegistryClient handles communication with the npm registry
type RegistryClient struct {
	baseURL     string
	httpClient  *http.Client
	retryConfig RetryConfig
	log         *logger.Logger
}

// VersionInfo contains metadata about a specific package version
type VersionInfo struct {
	Version      string            `json:"version"`
	Dependencies map[string]string `json:"dependencies,omitempty"`
	Dist         struct {
		Tarball string `json:"tarball"`
		Shasum  string `json:"shasum"`
	} `json:"dist"`
}

// PackageMetadata represents the npm package metadata
type PackageMetadata struct {
	Name     string                 `json:"name"`
	Versions map[string]VersionInfo `json:"versions"`
	DistTags map[string]string      `json:"dist-tags"`
}

// NewRegistryClient creates a new registry client
func NewRegistryClient(log *logger.Logger) *RegistryClient {
	return &RegistryClient{
		baseURL: defaultRegistryURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		retryConfig: RetryConfig{
			MaxRetries:  defaultRetires,
			RetryDelay:  time.Second,
			MaxWaitTime: time.Minute,
		},
		log: log,
	}
}

// GetPackageMetadata fetches metadata for a package
func (c *RegistryClient) GetPackageMetadata(name string) (*PackageMetadata, error) {
	url := fmt.Sprintf("%s/%s", c.baseURL, name)

	var metadata PackageMetadata
	err := c.fetchWithRetry(url, &metadata)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to fetch metadata for package %s", name))
	}

	return &metadata, nil
}

// GetPackageVersion fetches metadata for a specific version of a package
func (c *RegistryClient) GetPackageVersion(name, versionConstraint string) (*VersionInfo, error) {
	// First resolve the version constraint to a specific version
	version, err := c.resolveVersion(name, versionConstraint)
	if err != nil {
		return nil, err
	}

	metadata, err := c.GetPackageMetadata(name)
	if err != nil {
		return nil, err
	}

	versionInfo, exists := metadata.Versions[version]
	if !exists {
		return nil, errors.New("version_not_found", fmt.Sprintf("version %s not found for package %s", version, name), nil)
	}

	return &versionInfo, nil
}

// GetLatestVersion fetches the latest version of a package
func (c *RegistryClient) GetLatestVersion(name string) (*VersionInfo, error) {
	metadata, err := c.GetPackageMetadata(name)
	if err != nil {
		return nil, err
	}

	latestVersion, exists := metadata.DistTags["latest"]
	if !exists {
		return nil, errors.New("latest_version_not_found", fmt.Sprintf("latest version not found for package %s", name), nil)
	}

	return c.GetPackageVersion(name, latestVersion)
}

func (c *RegistryClient) SetBaseURL(url string) {
	c.baseURL = url
}

// fetchWithRetry performs an HTTP GET request with retry logic
func (c *RegistryClient) fetchWithRetry(url string, target interface{}) error {
	var lastErr error

	for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(c.retryConfig.RetryDelay * time.Duration(attempt))
		}

		resp, err := c.httpClient.Get(url)
		if err != nil {
			lastErr = err
			c.log.Warnf("Request failed (attempt %d): %v", attempt+1, err)
			continue
		}
		defer resp.Body.Close()

		// Handle rate limiting
		if resp.StatusCode == http.StatusTooManyRequests {
			c.log.Warn("Rate limited by registry")
			lastErr = errors.New("rate_limited", "too many requests to registry", nil)
			continue
		}

		// Handle other error status codes
		if resp.StatusCode >= 400 {
			body, _ := io.ReadAll(resp.Body)
			lastErr = errors.New("http_error", fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)), nil)
			continue
		}

		// Parse response
		if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
			lastErr = errors.New("parse_error", "failed to parse registry response", err)
			continue
		}

		return nil
	}

	return lastErr
}
