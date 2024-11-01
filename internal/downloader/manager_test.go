package downloader

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/marpit19/zap-pm/internal/logger"
	"github.com/marpit19/zap-pm/internal/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockTestServer represents our test server with necessary context
type MockTestServer struct {
	Server    *httptest.Server
	MockFiles map[string]MockFile
	BaseURL   string
}

// MockFile represents a test package file
type MockFile struct {
	Name    string
	Content string
	Shasum  string
}

// createMockFile creates a mock package file and returns its SHA1 hash
func createMockFile(content string) MockFile {
	hash := sha1.New()
	hash.Write([]byte(content))
	return MockFile{
		Name:    "package.tgz",
		Content: content,
		Shasum:  hex.EncodeToString(hash.Sum(nil)),
	}
}

// setupMockServer creates and configures the test server
func setupMockServer() *MockTestServer {
	// Create mock files
	mockFiles := map[string]MockFile{
		"express-4.17.1":     createMockFile("express-4.17.1-content"),
		"body-parser-1.19.0": createMockFile("body-parser-1.19.0-content"),
		"cookie-0.4.0":       createMockFile("cookie-0.4.0-content"),
	}

	mockServer := &MockTestServer{
		MockFiles: mockFiles,
	}

	// Create the test server
	mockServer.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Test server received request: %s\n", r.URL.Path)

		// Handle tarball downloads
		if strings.Contains(r.URL.Path, ".tgz") {
			mockServer.handleTarballDownload(w, r)
			return
		}

		// Handle metadata requests
		mockServer.handleMetadataRequest(w, r)
	}))

	mockServer.BaseURL = mockServer.Server.URL
	return mockServer
}

func (ms *MockTestServer) handleTarballDownload(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	packageName := pathParts[1]
	version := strings.TrimSuffix(pathParts[3], ".tgz")
	mockKey := fmt.Sprintf("%s-%s", packageName, version)

	if mockFile, ok := ms.MockFiles[mockKey]; ok {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write([]byte(mockFile.Content))
		return
	}

	w.WriteHeader(http.StatusNotFound)
}

func (ms *MockTestServer) handleMetadataRequest(w http.ResponseWriter, r *http.Request) {
	packageName := strings.Trim(r.URL.Path, "/")
	metadata := ms.createPackageMetadata(packageName)

	if metadata != "" {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, metadata)
		return
	}

	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, `{"error": "package not found"}`)
}

func (ms *MockTestServer) createPackageMetadata(packageName string) string {
	switch packageName {
	case "express":
		return fmt.Sprintf(`{
			"name": "express",
			"versions": {
				"4.17.1": {
					"version": "4.17.1",
					"dependencies": {
						"body-parser": "1.19.0",
						"cookie": "0.4.0"
					},
					"dist": {
						"tarball": "%s/express/-/4.17.1.tgz",
						"shasum": "%s"
					}
				}
			},
			"dist-tags": {
				"latest": "4.17.1"
			}
		}`, ms.BaseURL, ms.MockFiles["express-4.17.1"].Shasum)

	case "body-parser":
		return fmt.Sprintf(`{
			"name": "body-parser",
			"versions": {
				"1.19.0": {
					"version": "1.19.0",
					"dependencies": {},
					"dist": {
						"tarball": "%s/body-parser/-/1.19.0.tgz",
						"shasum": "%s"
					}
				}
			},
			"dist-tags": {
				"latest": "1.19.0"
			}
		}`, ms.BaseURL, ms.MockFiles["body-parser-1.19.0"].Shasum)

	case "cookie":
		return fmt.Sprintf(`{
			"name": "cookie",
			"versions": {
				"0.4.0": {
					"version": "0.4.0",
					"dependencies": {},
					"dist": {
						"tarball": "%s/cookie/-/0.4.0.tgz",
						"shasum": "%s"
					}
				}
			},
			"dist-tags": {
				"latest": "0.4.0"
			}
		}`, ms.BaseURL, ms.MockFiles["cookie-0.4.0"].Shasum)
	}

	return ""
}

// setupTestServer creates the test environment
func setupTestServer() (*MockTestServer, *registry.RegistryClient, *DownloadManager, string) {
	log := logger.New()

	// Setup mock server
	mockServer := setupMockServer()

	// Setup registry client
	registryClient := registry.NewRegistryClient(log)
	registryClient.SetBaseURL(mockServer.BaseURL)

	// Create temp directory for cache
	tempDir, _ := os.MkdirTemp("", "zap-test-*")

	// Create download manager
	dm := NewDownloadManager(registryClient, tempDir, log)

	return mockServer, registryClient, dm, tempDir
}

func TestDownloadPackage(t *testing.T) {
	mockServer, _, dm, tempDir := setupTestServer()
	defer mockServer.Server.Close()
	defer os.RemoveAll(tempDir)

	opts := DownloadOptions{
		UseCache:     true,
		ShowProgress: true,
	}

	// Test successful download
	result, err := dm.DownloadPackage("express", "4.17.1", opts)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify file exists and content
	assert.FileExists(t, result.Path)
	content, err := os.ReadFile(result.Path)
	require.NoError(t, err)
	assert.Equal(t, "express-4.17.1-content", string(content))

	// Test nonexistent package
	_, err = dm.DownloadPackage("nonexistent", "1.0.0", opts)
	assert.Error(t, err)
}

func TestDownloadDependencies(t *testing.T) {
	mockServer, _, dm, tempDir := setupTestServer()
	defer mockServer.Server.Close()
	defer os.RemoveAll(tempDir)

	opts := DownloadOptions{
		Concurrency:  2,
		UseCache:     true,
		ShowProgress: true,
	}

	results, err := dm.DownloadDependencies("express", "4.17.1", opts)
	require.NoError(t, err)
	assert.Len(t, results, 2) // body-parser and cookie

	// Verify dependencies
	for _, result := range results {
		assert.FileExists(t, result.Path)
		content, err := os.ReadFile(result.Path)
		require.NoError(t, err)
		assert.Contains(t, string(content), fmt.Sprintf("%s-%s-content", result.PackageName, result.Version))
	}
}

func TestCacheHandling(t *testing.T) {
	mockServer, _, dm, tempDir := setupTestServer()
	defer mockServer.Server.Close()
	defer os.RemoveAll(tempDir)

	opts := DownloadOptions{
		UseCache: true,
	}

	// First download
	start := time.Now()
	result1, err := dm.DownloadPackage("express", "4.17.1", opts)
	require.NoError(t, err)
	firstDownloadTime := time.Since(start)

	// Second download (should use cache)
	start = time.Now()
	result2, err := dm.DownloadPackage("express", "4.17.1", opts)
	require.NoError(t, err)
	cachedDownloadTime := time.Since(start)

	// Verify results
	assert.Equal(t, result1.Path, result2.Path)
	assert.True(t, cachedDownloadTime < firstDownloadTime)
}

func TestChecksumVerification(t *testing.T) {
	mockServer, _, dm, tempDir := setupTestServer()
	defer mockServer.Server.Close()
	defer os.RemoveAll(tempDir)

	// 1. First download - this should succeed
	result, err := dm.DownloadPackage("express", "4.17.1", DownloadOptions{UseCache: false})
	require.NoError(t, err)
	require.NotNil(t, result)

	// Store original path and shasum
	originalPath := result.Path
	originalShasum := result.Shasum

	// 2. Verify initial download
	assert.FileExists(t, originalPath)
	originalContent, err := os.ReadFile(originalPath)
	require.NoError(t, err)
	assert.NotEmpty(t, originalContent)

	// 3. Tamper with the file
	tamperedContent := []byte("deliberately corrupted content - this should cause a checksum mismatch")
	err = os.WriteFile(originalPath, tamperedContent, 0644)
	require.NoError(t, err)

	// Sleep briefly to ensure file write is complete
	time.Sleep(100 * time.Millisecond)

	// 4. Try to download again with cache enabled - this should detect the mismatch
	_, err = dm.DownloadPackage("express", "4.17.1", DownloadOptions{UseCache: true})
	require.Error(t, err, "Expected error due to checksum mismatch")
	assert.Contains(t, err.Error(), "checksum mismatch", "Error should mention checksum mismatch")

	// 5. Verify the tampered file was removed
	_, err = os.Stat(originalPath)
	assert.True(t, os.IsNotExist(err), "Tampered file should have been removed")

	// 6. Final download should succeed by fetching fresh copy
	finalResult, err := dm.DownloadPackage("express", "4.17.1", DownloadOptions{UseCache: true})
	require.NoError(t, err)
	require.NotNil(t, finalResult)

	// 7. Verify the content was restored correctly
	finalContent, err := os.ReadFile(finalResult.Path)
	require.NoError(t, err)
	assert.NotEmpty(t, finalContent)
	assert.Equal(t, originalShasum, finalResult.Shasum)

	// 8. Verify we can now use the cache successfully
	cachedResult, err := dm.DownloadPackage("express", "4.17.1", DownloadOptions{UseCache: true})
	require.NoError(t, err)
	require.NotNil(t, cachedResult)
	assert.Equal(t, finalResult.Path, cachedResult.Path)
}

func cleanup(t *testing.T, paths ...string) {
	for _, path := range paths {
		if path != "" {
			err := os.RemoveAll(path)
			assert.NoError(t, err, "cleanup failed for path: "+path)
		}
	}
}

func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()

	// Cleanup any leftover test directories
	pattern := filepath.Join(os.TempDir(), "zap-test-*")
	matches, err := filepath.Glob(pattern)
	if err == nil {
		for _, match := range matches {
			os.RemoveAll(match)
		}
	}

	os.Exit(code)
}
