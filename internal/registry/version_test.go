package registry

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/marpit19/zap-pm/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// First, let's modify setupTestServer to provide consistent versions for testing
func setupVersionTestServer() (*httptest.Server, *RegistryClient) {
	log := logger.New()
	client := NewRegistryClient(log)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/express":
			json.NewEncoder(w).Encode(PackageMetadata{
				Name: "express",
				Versions: map[string]VersionInfo{
					"4.17.0": {
						Version: "4.17.0",
						Dependencies: map[string]string{
							"body-parser": "1.19.0",
						},
					},
					"4.17.1": {
						Version: "4.17.1",
						Dependencies: map[string]string{
							"body-parser": "1.19.0",
						},
					},
				},
				DistTags: map[string]string{
					"latest": "4.17.1",
				},
			})
		case "/nonexistent":
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": "Not found"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	client.SetBaseURL(server.URL)
	return server, client
}

func TestVersionResolution(t *testing.T) {
	server, client := setupVersionTestServer()
	defer server.Close()

	tests := []struct {
		name        string
		pkgName     string
		constraint  string
		expected    string
		expectError bool
	}{
		{
			name:       "exact version",
			pkgName:    "express",
			constraint: "4.17.1",
			expected:   "4.17.1",
		},
		{
			name:       "caret range",
			pkgName:    "express",
			constraint: "^4.17.0",
			expected:   "4.17.1",
		},
		{
			name:       "tilde range",
			pkgName:    "express",
			constraint: "~4.17.0",
			expected:   "4.17.1",
		},
		{
			name:       "greater than",
			pkgName:    "express",
			constraint: ">=4.17.0",
			expected:   "4.17.1",
		},
		{
			name:        "invalid constraint",
			pkgName:     "express",
			constraint:  "invalid",
			expectError: true,
		},
		{
			name:        "no matching version",
			pkgName:     "express",
			constraint:  "^99.0.0",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := client.resolveVersion(tt.pkgName, tt.constraint)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, version)

			// Verify we can actually get this version
			versionInfo, err := client.GetPackageVersion(tt.pkgName, version)
			require.NoError(t, err)
			assert.Equal(t, version, versionInfo.Version)
		})
	}
}

func TestIsExactVersion(t *testing.T) {
	tests := []struct {
		version string
		want    bool
	}{
		{"1.0.0", true},
		{"1.2.3-alpha.1", true},
		{"^1.0.0", false},
		{"~1.0.0", false},
		{">1.0.0", false},
		{">=1.0.0", false},
		{"latest", false},
		{"*", false},
		{"", false},
		{" 1.0.0 ", true},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			got := isExactVersion(tt.version)
			assert.Equal(t, tt.want, got)
		})
	}
}
