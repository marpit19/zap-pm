package registry

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/marpit19/zap-pm/internal/logger"
	"github.com/stretchr/testify/assert"
)

func setupTestServer() (*httptest.Server, *RegistryClient) {
	log := logger.New()
	client := NewRegistryClient(log)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/express":
			json.NewEncoder(w).Encode(PackageMetadata{
				Name: "express",
				Versions: map[string]VersionInfo{
					"4.17.1": {
						Version: "4.17.1",
						Dependencies: map[string]string{
							"body-parser": "1.19.0",
							"cookie":      "0.4.0",
						},
						Dist: struct {
							Tarball string "json:\"tarball\""
							Shasum  string "json:\"shasum\""
						}{
							Tarball: "https://registry.npmjs.org/express/-/express-4.17.1.tgz",
							Shasum:  "4491fc38605cf51f8629d39c2b5d026f98a4c134",
						},
					},
					"4.17.2": {
						Version: "4.17.2",
						Dependencies: map[string]string{
							"body-parser": "1.19.1",
							"cookie":      "0.4.1",
						},
					},
				},
				DistTags: map[string]string{
					"latest": "4.17.2",
				},
			})
		case "/nonexistent":
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": "Not found"}`))
		case "/ratelimit":
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error": "Too many requests"}`))
		case "/malformed":
			w.Write([]byte(`{invalid json`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	client.baseURL = server.URL
	client.retryConfig.RetryDelay = time.Millisecond // Speed up tests
	return server, client
}

func TestGetPackageMetadata(t *testing.T) {
	server, client := setupTestServer()
	defer server.Close()

	tests := []struct {
		name          string
		packageName   string
		expectError   bool
		errorContains string
		validate      func(*PackageMetadata)
	}{
		{
			name:        "valid package",
			packageName: "express",
			validate: func(meta *PackageMetadata) {
				assert.Equal(t, "express", meta.Name)
				assert.Equal(t, 2, len(meta.Versions))
				assert.Equal(t, "4.17.2", meta.DistTags["latest"])
			},
		},
		{
			name:          "nonexistent package",
			packageName:   "nonexistent",
			expectError:   true,
			errorContains: "HTTP 404",
		},
		{
			name:          "rate limited",
			packageName:   "ratelimit",
			expectError:   true,
			errorContains: "too many requests",
		},
		{
			name:          "malformed response",
			packageName:   "malformed",
			expectError:   true,
			errorContains: "failed to parse registry response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta, err := client.GetPackageMetadata(tt.packageName)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, meta)
				if tt.validate != nil {
					tt.validate(meta)
				}
			}
		})
	}
}

func TestGetPackageVersion(t *testing.T) {
	server, client := setupTestServer()
	defer server.Close()

	tests := []struct {
		name          string
		packageName   string
		version       string
		expectError   bool
		errorContains string
		validate      func(*VersionInfo)
	}{
		{
			name:        "specific version",
			packageName: "express",
			version:     "4.17.1",
			validate: func(info *VersionInfo) {
				assert.Equal(t, "4.17.1", info.Version)
				assert.Equal(t, "1.19.0", info.Dependencies["body-parser"])
				assert.Equal(t, "0.4.0", info.Dependencies["cookie"])
				assert.NotEmpty(t, info.Dist.Tarball)
				assert.NotEmpty(t, info.Dist.Shasum)
			},
		},
		{
			name:          "nonexistent version",
			packageName:   "express",
			version:       "0.0.1",
			expectError:   true,
			errorContains: "version_not_found", // Updated to match actual error type
		},
		{
			name:          "nonexistent package",
			packageName:   "nonexistent",
			version:       "1.0.0",
			expectError:   true,
			errorContains: "HTTP 404",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := client.GetPackageVersion(tt.packageName, tt.version)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, info)
				if tt.validate != nil {
					tt.validate(info)
				}
			}
		})
	}
}

func TestGetLatestVersion(t *testing.T) {
	server, client := setupTestServer()
	defer server.Close()

	tests := []struct {
		name          string
		packageName   string
		expectError   bool
		errorContains string
		validate      func(*VersionInfo)
	}{
		{
			name:        "valid package",
			packageName: "express",
			validate: func(info *VersionInfo) {
				assert.Equal(t, "4.17.2", info.Version)
				assert.Equal(t, "1.19.1", info.Dependencies["body-parser"])
				assert.Equal(t, "0.4.1", info.Dependencies["cookie"])
			},
		},
		{
			name:          "nonexistent package",
			packageName:   "nonexistent",
			expectError:   true,
			errorContains: "HTTP 404",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := client.GetLatestVersion(tt.packageName)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, info)
				if tt.validate != nil {
					tt.validate(info)
				}
			}
		})
	}
}

func TestRegistryClientRetry(t *testing.T) {
	log := logger.New()
	client := NewRegistryClient(log)

	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts <= 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		json.NewEncoder(w).Encode(PackageMetadata{
			Name: "test-package",
			Versions: map[string]VersionInfo{
				"1.0.0": {Version: "1.0.0"},
			},
		})
	}))
	defer server.Close()

	client.baseURL = server.URL
	client.retryConfig.RetryDelay = time.Millisecond

	meta, err := client.GetPackageMetadata("test-package")
	assert.NoError(t, err)
	assert.NotNil(t, meta)
	assert.Equal(t, 3, attempts)
}
