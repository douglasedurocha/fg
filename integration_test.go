package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fhir-guard/fg/cmd"
	"github.com/fhir-guard/fg/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegrationCommandExecution tests the full command execution flow
func TestIntegrationCommandExecution(t *testing.T) {
	// Skip in CI environments or if integration tests are disabled
	if os.Getenv("SKIP_INTEGRATION_TESTS") != "" {
		t.Skip("Skipping integration tests")
	}

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "fg-integration-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Set up a mock server to serve version information and downloads
	mockServer := setupMockServer(t)
	defer mockServer.Close()

	// Set up the environment with the test directory as FG_HOME
	os.Setenv("FG_HOME", tempDir)
	defer os.Unsetenv("FG_HOME")

	// Prepare a custom configuration pointing to our mock server
	configPath := filepath.Join(tempDir, "config", "config.yaml")
	err = os.MkdirAll(filepath.Join(tempDir, "config"), 0755)
	require.NoError(t, err)

	cfg := config.DefaultConfig()
	cfg.FGHome = tempDir
	cfg.DownloadURL = mockServer.URL
	err = config.SaveConfig(cfg)
	require.NoError(t, err)

	// Test the command execution flow by capturing output
	testCases := []struct {
		name    string
		args    []string
		verify  func(t *testing.T, output string, err error)
	}{
		{
			name: "List command with no installations",
			args: []string{"list"},
			verify: func(t *testing.T, output string, err error) {
				assert.NoError(t, err)
				assert.Contains(t, output, "Nenhuma versão instalada")
			},
		},
		{
			name: "List remote versions",
			args: []string{"list", "--remote"},
			verify: func(t *testing.T, output string, err error) {
				assert.NoError(t, err)
				assert.Contains(t, output, "1.0.0")
				assert.Contains(t, output, "1.1.0")
			},
		},
		{
			name: "Install command",
			args: []string{"install", "1.0.0"},
			verify: func(t *testing.T, output string, err error) {
				assert.NoError(t, err)
				assert.Contains(t, output, "FHIR Guard versão 1.0.0 instalada com sucesso")
				
				// Verify files were created
				jarPath := filepath.Join(tempDir, "versions", "1.0.0", "fhir-guard-1.0.0.jar")
				_, err := os.Stat(jarPath)
				assert.NoError(t, err, "JAR file should exist")
			},
		},
		{
			name: "List installed versions",
			args: []string{"list"},
			verify: func(t *testing.T, output string, err error) {
				assert.NoError(t, err)
				assert.Contains(t, output, "1.0.0")
				assert.Contains(t, output, "instalada")
			},
		},
		{
			name: "Get version information",
			args: []string{"version"},
			verify: func(t *testing.T, output string, err error) {
				assert.NoError(t, err)
				// The version output depends on your application
				assert.Contains(t, output, "FHIR Guard CLI")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := executeCommand(cmd.RootCmd, tc.args...)
			tc.verify(t, output, err)
		})
	}
}

// executeCommand is a helper function that executes a cobra command with the given arguments
func executeCommand(root *cobra.Command, args ...string) (string, error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	
	// Create a new command to avoid side effects
	cmd := *root
	cmd.SetArgs(args)
	err := cmd.Execute()
	
	return buf.String(), err
}

// setupMockServer creates a test HTTP server that responds with mock data
func setupMockServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/versions/index.json"):
			// Serve the index of versions
			versions := []map[string]interface{}{
				{
					"version":     "1.0.0",
					"releaseDate": "2023-01-01T00:00:00Z",
				},
				{
					"version":     "1.1.0",
					"releaseDate": "2023-02-01T00:00:00Z",
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(versions)
			
		case strings.Contains(r.URL.Path, "/versions/1.0.0/metadata.json"):
			// Serve metadata for version 1.0.0
			metadata := map[string]interface{}{
				"url":          fmt.Sprintf("%s/downloads/fhir-guard-1.0.0.jar", r.Host),
				"checksum":     "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", // Empty file checksum
				"dependencies": []string{},
				"requiredJava": "11",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(metadata)
			
		case strings.Contains(r.URL.Path, "/downloads/fhir-guard-1.0.0.jar"):
			// Serve an empty JAR file
			w.Header().Set("Content-Type", "application/java-archive")
			w.Write([]byte("mock jar file content"))
			
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

// TestEndToEndUserFlow simulates a real-world scenario where a user:
// 1. Lists available versions
// 2. Installs a specific version
// 3. Verifies the installation
func TestEndToEndUserFlow(t *testing.T) {
	// Skip in CI environments or if e2e tests are disabled
	if os.Getenv("SKIP_E2E_TESTS") != "" {
		t.Skip("Skipping end-to-end tests")
	}

	// Build the binary for testing
	tempBinDir, err := os.MkdirTemp("", "fg-e2e-bin")
	require.NoError(t, err)
	defer os.RemoveAll(tempBinDir)

	binaryPath := filepath.Join(tempBinDir, "fg")
	cmd := exec.Command("go", "build", "-o", binaryPath)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to build binary: %s", string(output))

	// Set up a test directory as FG_HOME
	testHome, err := os.MkdirTemp("", "fg-e2e-home")
	require.NoError(t, err)
	defer os.RemoveAll(testHome)

	// Set up a mock server
	mockServer := setupMockServer(t)
	defer mockServer.Close()

	// Create a test config file pointing to our mock server
	configDir := filepath.Join(testHome, "config")
	err = os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	cfg := config.DefaultConfig()
	cfg.FGHome = testHome
	cfg.DownloadURL = mockServer.URL
	cfgBytes, err := json.Marshal(cfg)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(configDir, "config.json"), cfgBytes, 0644)
	require.NoError(t, err)

	// Run commands in sequence to simulate user flow
	testEnv := []string{fmt.Sprintf("FG_HOME=%s", testHome)}
	
	// Step 1: List versions (should be empty initially)
	listCmd := exec.Command(binaryPath, "list")
	listCmd.Env = append(os.Environ(), testEnv...)
	listOutput, err := listCmd.Output()
	assert.NoError(t, err)
	assert.Contains(t, string(listOutput), "Nenhuma versão instalada")

	// Step 2: List remote versions
	listRemoteCmd := exec.Command(binaryPath, "list", "--remote")
	listRemoteCmd.Env = append(os.Environ(), testEnv...)
	listRemoteOutput, err := listRemoteCmd.Output()
	assert.NoError(t, err)
	assert.Contains(t, string(listRemoteOutput), "1.0.0")

	// Step 3: Install a version
	installCmd := exec.Command(binaryPath, "install", "1.0.0")
	installCmd.Env = append(os.Environ(), testEnv...)
	installOutput, err := installCmd.CombinedOutput()
	assert.NoError(t, err, "Installation failed: %s", string(installOutput))
	assert.Contains(t, string(installOutput), "instalada com sucesso")

	// Step 4: Verify installation
	verifyCmd := exec.Command(binaryPath, "list")
	verifyCmd.Env = append(os.Environ(), testEnv...)
	verifyOutput, err := verifyCmd.Output()
	assert.NoError(t, err)
	assert.Contains(t, string(verifyOutput), "1.0.0")
	assert.Contains(t, string(verifyOutput), "instalada")

	// Verify actual file exists
	jarPath := filepath.Join(testHome, "versions", "1.0.0", "fhir-guard-1.0.0.jar")
	_, err = os.Stat(jarPath)
	assert.NoError(t, err, "JAR file should exist")
} 