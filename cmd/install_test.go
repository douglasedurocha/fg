package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fhir-guard/fg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateVersion(t *testing.T) {
	tests := []struct {
		version string
		valid   bool
	}{
		{"1.0.0", true},
		{"11.2.3", true},
		{"1.0", false},
		{"1", false},
		{"v1.0.0", false},
		{"1.0.0-alpha", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			result := validateVersion(tt.version)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestVerifyChecksum(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test-file-*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Write some content with explicitly using binary mode to avoid line ending issues
	content := []byte("test content for checksum verification")
	err = os.WriteFile(tmpFile.Name(), content, 0644)
	require.NoError(t, err)

	// Compute SHA-256 checksum of the exact bytes we wrote
	hasher := sha256.New()
	hasher.Write(content)
	expectedHash := hex.EncodeToString(hasher.Sum(nil))

	// Test valid checksum
	err = verifyChecksum(tmpFile.Name(), expectedHash)
	assert.NoError(t, err)

	// Test invalid checksum
	err = verifyChecksum(tmpFile.Name(), "invalid-checksum")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "checksum inválido")
}

func TestDownloadFile(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "success") {
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, "downloaded content")
		} else if strings.Contains(r.URL.Path, "notfound") {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "download-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Test successful download
	successFilePath := filepath.Join(tmpDir, "success.txt")
	err = downloadFile(server.URL+"/success", successFilePath)
	assert.NoError(t, err)

	content, err := os.ReadFile(successFilePath)
	require.NoError(t, err)
	assert.Equal(t, "downloaded content", string(content))

	// Test not found error
	notFoundPath := filepath.Join(tmpDir, "notfound.txt")
	err = downloadFile(server.URL+"/notfound", notFoundPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "download falhou")

	// Test server error
	serverErrorPath := filepath.Join(tmpDir, "error.txt")
	err = downloadFile(server.URL+"/error", serverErrorPath)
	assert.Error(t, err)
}

func TestFetchVersionInfo(t *testing.T) {
	// Create a test server with mock version metadata
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "versions/1.0.0/metadata.json") {
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, `{
				"url": "https://example.com/downloads/fhir-guard-1.0.0.jar",
				"checksum": "abc123",
				"dependencies": ["https://example.com/lib1.jar", "https://example.com/lib2.jar"],
				"requiredJava": "11"
			}`)
		} else if strings.Contains(r.URL.Path, "versions/9.9.9/metadata.json") {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create test config
	cfg := &config.FGConfig{
		DownloadURL: server.URL,
		Versions:    make(map[string]config.VersionInfo),
	}

	// Test successful fetch
	info, err := fetchVersionInfo(cfg, "1.0.0")
	assert.NoError(t, err)
	assert.Equal(t, "https://example.com/downloads/fhir-guard-1.0.0.jar", info.URL)
	assert.Equal(t, "abc123", info.Checksum)
	assert.Equal(t, 2, len(info.Dependencies))
	assert.Equal(t, "11", info.RequiredJava)

	// Test version not found
	_, err = fetchVersionInfo(cfg, "9.9.9")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "versão 9.9.9 não encontrada")

	// Test with preexisting version in config
	cfg.Versions["2.0.0"] = config.VersionInfo{
		URL:      "https://example.com/downloads/fhir-guard-2.0.0.jar",
		Checksum: "def456",
	}
	info, err = fetchVersionInfo(cfg, "2.0.0")
	assert.NoError(t, err)
	assert.Equal(t, "https://example.com/downloads/fhir-guard-2.0.0.jar", info.URL)
	assert.Equal(t, "def456", info.Checksum)
} 