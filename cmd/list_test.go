package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fhir-guard/fg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.0", "1.0.1", -1},
		{"1.0.1", "1.0.0", 1},
		{"1.0.0", "1.1.0", -1},
		{"2.0.0", "1.9.9", 1},
		{"1.10.0", "1.9.0", 1},
		{"1.0.10", "1.0.9", 1},
	}

	for _, tt := range tests {
		t.Run(tt.v1+" vs "+tt.v2, func(t *testing.T) {
			result := compareVersions(tt.v1, tt.v2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetInstalledVersions(t *testing.T) {
	// Create temp directory for testing
	tempDir, err := os.MkdirTemp("", "fg-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test config
	cfg := &config.FGConfig{
		FGHome: tempDir,
	}

	// Create versions directory structure
	versionsDir := filepath.Join(tempDir, "versions")
	require.NoError(t, os.MkdirAll(versionsDir, 0755))

	// No versions installed initially
	versions := getInstalledVersions(cfg)
	assert.Empty(t, versions)

	// Create some fake installed versions
	testVersions := []string{"1.0.0", "1.1.0", "2.0.0"}
	for _, v := range testVersions {
		versionDir := filepath.Join(versionsDir, v)
		require.NoError(t, os.MkdirAll(versionDir, 0755))
		jarPath := filepath.Join(versionDir, "fhir-guard-"+v+".jar")
		require.NoError(t, os.WriteFile(jarPath, []byte("fake jar content"), 0644))
	}

	// Create a directory without jar file (should be ignored)
	emptyDir := filepath.Join(versionsDir, "empty")
	require.NoError(t, os.MkdirAll(emptyDir, 0755))

	// Test getting installed versions
	versions = getInstalledVersions(cfg)
	assert.Len(t, versions, 3)

	// Verify versions
	versionMap := make(map[string]bool)
	for _, v := range versions {
		versionMap[v.Version] = true
		assert.True(t, v.Installed)
	}

	for _, v := range testVersions {
		assert.True(t, versionMap[v], "Expected version %s to be installed", v)
	}
}

func TestGetRemoteVersions(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/versions/index.json" {
			releases := []VersionMeta{
				{
					Version:     "1.0.0",
					ReleaseDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				},
				{
					Version:     "1.1.0",
					ReleaseDate: time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC),
				},
				{
					Version:     "2.0.0",
					ReleaseDate: time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC),
				},
			}
			json.NewEncoder(w).Encode(releases)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create test config
	cfg := &config.FGConfig{
		DownloadURL: server.URL,
	}

	// Test getting remote versions
	versions, err := getRemoteVersions(cfg)
	assert.NoError(t, err)
	assert.Len(t, versions, 3)

	// Verify versions are sorted correctly and latest is marked
	assert.Equal(t, "2.0.0", versions[0].Version)
	assert.True(t, versions[0].IsLatest)
	assert.False(t, versions[1].IsLatest)
	assert.False(t, versions[2].IsLatest)

	// Test with a server that returns an error
	cfg.DownloadURL = "http://non-existent-server"
	_, err = getRemoteVersions(cfg)
	assert.Error(t, err)
}

func TestGetRemoteVersionsServerError(t *testing.T) {
	// Create a server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/versions/error-404/index.json" {
			w.WriteHeader(http.StatusNotFound)
		} else if r.URL.Path == "/versions/error-500/index.json" {
			w.WriteHeader(http.StatusInternalServerError)
		} else if r.URL.Path == "/versions/bad-json/index.json" {
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, "This is not valid JSON")
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Test 404 error
	cfg := &config.FGConfig{
		DownloadURL: server.URL + "/versions/error-404",
	}
	_, err := getRemoteVersions(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao obter índice de versões")

	// Test 500 error
	cfg.DownloadURL = server.URL + "/versions/error-500"
	_, err = getRemoteVersions(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao obter índice de versões")

	// Test invalid JSON
	cfg.DownloadURL = server.URL + "/versions/bad-json"
	_, err = getRemoteVersions(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao decodificar índice")
} 