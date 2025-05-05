package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitFGHome(t *testing.T) {
	// Test with FG_HOME environment variable
	tempDir, err := os.MkdirTemp("", "fg-test-home")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Salve o valor original para restaurá-lo depois
	originalFGHome := os.Getenv("FG_HOME")
	originalHome := os.Getenv("HOME")
	defer func() {
		os.Setenv("FG_HOME", originalFGHome)
		os.Setenv("HOME", originalHome)
	}()

	// Configure o ambiente para o teste
	os.Setenv("FG_HOME", tempDir)

	home, err := InitFGHome()
	assert.NoError(t, err)
	assert.Equal(t, tempDir, home)

	// Check if directories were created
	for _, dir := range []string{"versions", "logs", "config"} {
		dirPath := filepath.Join(tempDir, dir)
		info, err := os.Stat(dirPath)
		assert.NoError(t, err)
		assert.True(t, info.IsDir())
	}

	// Test with default home directory (in a temporary location)
	os.Unsetenv("FG_HOME")
	
	// No Windows, precisamos usar a variável USERPROFILE como HOME
	if os.Getenv("OS") == "Windows_NT" {
		os.Setenv("USERPROFILE", tempDir)
		defer os.Setenv("USERPROFILE", originalHome)
	} else {
		os.Setenv("HOME", tempDir)
	}

	home, err = InitFGHome()
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(tempDir, ".fhir-guard"), home)
}

func TestLoadAndSaveConfig(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "fg-config-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Set up environment for testing
	os.Setenv("FG_HOME", tempDir)
	defer os.Unsetenv("FG_HOME")

	// Initialize home directory
	_, err = InitFGHome()
	require.NoError(t, err)

	// Create a test config
	testConfig := &FGConfig{
		FGHome:      tempDir,
		DownloadURL: "https://test-download-url.com",
		LogLevel:    "debug",
		Java: JavaConfig{
			MinVersion: "11",
			MaxVersion: "17",
		},
		Versions: map[string]VersionInfo{
			"1.0.0": {
				URL:          "https://example.com/jdk-1.0.0.jar",
				Checksum:     "abc123",
				RequiredJava: "11",
				Installed:    true,
			},
		},
		ActivePIDs: map[string]int{
			"test-instance": 12345,
		},
	}

	// Save the config
	err = SaveConfig(testConfig)
	assert.NoError(t, err)

	// Load the config
	loadedConfig, err := LoadConfig()
	assert.NoError(t, err)

	// Verify config values (exclude ActivePIDs as it's stored separately)
	assert.Equal(t, testConfig.DownloadURL, loadedConfig.DownloadURL)
	assert.Equal(t, testConfig.LogLevel, loadedConfig.LogLevel)
	assert.Equal(t, testConfig.Java.MinVersion, loadedConfig.Java.MinVersion)
	assert.Equal(t, testConfig.Java.MaxVersion, loadedConfig.Java.MaxVersion)
	assert.Contains(t, loadedConfig.Versions, "1.0.0")
	assert.Equal(t, testConfig.Versions["1.0.0"].URL, loadedConfig.Versions["1.0.0"].URL)
	assert.Equal(t, testConfig.Versions["1.0.0"].Checksum, loadedConfig.Versions["1.0.0"].Checksum)
	assert.Equal(t, testConfig.Versions["1.0.0"].RequiredJava, loadedConfig.Versions["1.0.0"].RequiredJava)
	assert.Equal(t, testConfig.Versions["1.0.0"].Installed, loadedConfig.Versions["1.0.0"].Installed)
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	
	// Test default values
	assert.NotNil(t, cfg)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "https://releases.fhir-guard.org", cfg.DownloadURL)
	assert.NotNil(t, cfg.Versions)
	assert.Empty(t, cfg.Versions)
	
	// Test Java config
	assert.Contains(t, cfg.Java.JvmArgs, "-Xms256m")
	assert.Contains(t, cfg.Java.JvmArgs, "-Xmx1g")
	
	// Test Server config
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "1g", cfg.Server.MaxMemory)
	assert.Contains(t, cfg.Server.Contexts, "/fhir")
	assert.NotNil(t, cfg.Server.Env)
} 