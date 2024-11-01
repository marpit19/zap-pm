package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionResolution(t *testing.T) {
    server, client := setupTestServer()
    defer server.Close()

    tests := []struct {
        name          string
        pkgName       string
        constraint    string
        expected     string
        expectError  bool
    }{
        {
            name:       "exact version",
            pkgName:    "express",
            constraint: "4.17.1",
            expected:  "4.17.1",
        },
        {
            name:       "caret range",
            pkgName:    "express",
            constraint: "^4.17.0",
            expected:  "4.17.1",
        },
        {
            name:       "tilde range",
            pkgName:    "express",
            constraint: "~4.17.0",
            expected:  "4.17.1",
        },
        {
            name:       "greater than",
            pkgName:    "express",
            constraint: ">=4.17.0",
            expected:  "4.17.1",
        },
        {
            name:       "invalid constraint",
            pkgName:    "express",
            constraint: "invalid",
            expectError: true,
        },
        {
            name:       "no matching version",
            pkgName:    "express",
            constraint: "^99.0.0",
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
