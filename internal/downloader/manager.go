package downloader

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/marpit19/zap-pm/internal/logger"
	"github.com/marpit19/zap-pm/internal/registry"
)

const (
	defaultConcurrency = 3
	defaultTimeout     = 30 * time.Second
	defaultBufferSize  = 32 * 1024 // 32KB buffer
)

// DownloadOptions configures the download behavior
type DownloadOptions struct {
	Concurrency  int
	UseCache     bool
	ShowProgress bool
	Timeout      time.Duration
}

// DownloadResult contains information about a completed download
type DownloadResult struct {
	PackageName string
	Version     string
	Path        string
	Shasum      string
	Error       error
}

// DownloadManager handles package downloads
type DownloadManager struct {
	client      *http.Client
	registry    *registry.RegistryClient
	cacheDir    string
	log         *logger.Logger
	progressBar *ProgressBar
	mu          sync.Mutex
}

// NewDownloadManager creates a new download manager
func NewDownloadManager(registryClient *registry.RegistryClient, cacheDir string, log *logger.Logger) *DownloadManager {
	return &DownloadManager{
		client: &http.Client{
			Timeout: defaultTimeout,
		},
		registry: registryClient,
		cacheDir: cacheDir,
		log:      log,
	}
}

// DownloadPackage downloads a single package
func (dm *DownloadManager) DownloadPackage(name, version string, opts DownloadOptions) (*DownloadResult, error) {
	dm.log.Infof("Downloading package %s@%s", name, version)

	// Get package metadata
	versionInfo, err := dm.registry.GetPackageVersion(name, version)
	if err != nil {
		return nil, fmt.Errorf("failed to get package metadata: %w", err)
	}

	// Check cache first if enabled
	if opts.UseCache {
		if cachedPath, exists, err := dm.checkCache(name, version, versionInfo.Dist.Shasum); err != nil {
			// Propagate checksum mismatch error
			return nil, fmt.Errorf("cache validation failed: %w", err)
		} else if exists {
			dm.log.Infof("Using cached version from: %s", cachedPath)
			return &DownloadResult{
				PackageName: name,
				Version:     version,
				Path:        cachedPath,
				Shasum:      versionInfo.Dist.Shasum,
			}, nil
		}
	}

	// Cache miss or disabled, download the package
	targetDir := filepath.Join(dm.cacheDir, name, version)
	targetPath := filepath.Join(targetDir, "package.tgz")

	// Download the package
	if err := dm.downloadFile(versionInfo.Dist.Tarball, targetPath, versionInfo.Dist.Shasum, opts.ShowProgress); err != nil {
		return nil, err
	}

	return &DownloadResult{
		PackageName: name,
		Version:     version,
		Path:        targetPath,
		Shasum:      versionInfo.Dist.Shasum,
	}, nil
}

// DownloadDependencies downloads all dependencies for a package
func (dm *DownloadManager) DownloadDependencies(name, version string, opts DownloadOptions) ([]*DownloadResult, error) {
	dm.log.Infof("Downloading dependencies for %s@%s", name, version)

	// Get package metadata
	versionInfo, err := dm.registry.GetPackageVersion(name, version)
	if err != nil {
		return nil, fmt.Errorf("failed to get package metadata: %w", err)
	}

	if len(versionInfo.Dependencies) == 0 {
		dm.log.Info("No dependencies found")
		return []*DownloadResult{}, nil
	}

	dm.log.Infof("Found %d dependencies", len(versionInfo.Dependencies))

	// Create work queue for dependencies
	var wg sync.WaitGroup
	results := make([]*DownloadResult, 0)
	resultsChan := make(chan *DownloadResult, len(versionInfo.Dependencies))
	errorsChan := make(chan error, len(versionInfo.Dependencies))

	// Set concurrency
	if opts.Concurrency <= 0 {
		opts.Concurrency = defaultConcurrency
	}

	// Create worker pool
	semaphore := make(chan struct{}, opts.Concurrency)

	// Process each dependency
	for depName, depVersion := range versionInfo.Dependencies {
		wg.Add(1)
		go func(name, version string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			dm.log.Debugf("Downloading dependency: %s@%s", name, version)
			result, err := dm.DownloadPackage(name, version, opts)
			if err != nil {
				errorsChan <- fmt.Errorf("failed to download %s@%s: %w", name, version, err)
				return
			}
			resultsChan <- result
		}(depName, depVersion)
	}

	// Wait for all downloads to complete
	go func() {
		wg.Wait()
		close(resultsChan)
		close(errorsChan)
	}()

	// Collect results and errors
	var downloadErrors []error
	for err := range errorsChan {
		downloadErrors = append(downloadErrors, err)
	}
	for result := range resultsChan {
		results = append(results, result)
	}

	if len(downloadErrors) > 0 {
		return results, fmt.Errorf("some dependencies failed to download: %v", downloadErrors)
	}

	dm.log.Infof("Successfully downloaded %d dependencies", len(results))
	return results, nil
}

// downloadFile downloads a file and verifies its checksum
func (dm *DownloadManager) downloadFile(url, targetPath, expectedShasum string, showProgress bool) error {
	dm.log.Debugf("Downloading from URL: %s", url)
	dm.log.Debugf("Target path: %s", targetPath)

	// Create the target directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Create the target file
	out, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("failed to create target file: %w", err)
	}
	defer out.Close()

	// Create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add relevant headers
	req.Header.Set("Accept", "application/octet-stream")
	req.Header.Set("User-Agent", "zap-package-manager/0.1.0")

	// Get the file
	resp, err := dm.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file (status %d): %s", resp.StatusCode, resp.Status)
	}

	// Setup progress bar if requested
	var reader io.Reader = resp.Body
	if showProgress {
		dm.mu.Lock()
		dm.progressBar = NewProgressBar(resp.ContentLength)
		dm.mu.Unlock()
		reader = dm.progressBar.NewProxyReader(resp.Body)
	}

	// Setup checksum calculation
	hash := sha1.New()
	writer := io.MultiWriter(out, hash)

	// Copy the data
	written, err := io.Copy(writer, reader)
	if err != nil {
		os.Remove(targetPath)
		return fmt.Errorf("download interrupted: %w", err)
	}

	dm.log.Debugf("Downloaded %d bytes", written)

	// Verify checksum
	actualShasum := hex.EncodeToString(hash.Sum(nil))
	if actualShasum != expectedShasum {
		os.Remove(targetPath)
		return fmt.Errorf("checksum mismatch (expected: %s, got: %s)", expectedShasum, actualShasum)
	}

	dm.log.Debugf("Checksum verified: %s", actualShasum)
	return nil
}

// checkCache looks for a package in the cache
func (dm *DownloadManager) checkCache(name, version, expectedShasum string) (string, bool, error) {
	path := filepath.Join(dm.cacheDir, name, version, "package.tgz")
	dm.log.Debugf("Checking cache for %s@%s at %s", name, version, path)

	// Check if file exists
	if _, err := os.Stat(path); err != nil {
		dm.log.Debugf("Cache miss: file not found")
		return "", false, nil
	}

	// Read the entire file
	content, err := os.ReadFile(path)
	if err != nil {
		dm.log.Debugf("Cache miss: failed to read file: %v", err)
		return "", false, nil
	}

	// Calculate checksum
	hash := sha1.New()
	hash.Write(content)
	actualShasum := hex.EncodeToString(hash.Sum(nil))

	dm.log.Debugf("Cache checksum comparison - Expected: %s, Got: %s", expectedShasum, actualShasum)

	if actualShasum != expectedShasum {
		dm.log.Warn("Cache miss: checksum mismatch")
		// Remove invalid cache entry
		os.Remove(path)
		return "", false, fmt.Errorf("checksum mismatch in cached file (expected: %s, got: %s)", expectedShasum, actualShasum)
	}

	dm.log.Debug("Cache hit: checksums match")
	return path, true, nil
}
