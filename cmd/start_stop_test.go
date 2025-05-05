package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/fhir-guard/fg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsJavaVersionValid(t *testing.T) {
	tests := []struct {
		name         string
		javaVersion  string
		minVersion   string
		maxVersion   string
		expectedBool bool
	}{
		{
			name:         "Java version within range",
			javaVersion:  "11.0.1",
			minVersion:   "11",
			maxVersion:   "17",
			expectedBool: true,
		},
		{
			name:         "Java version below minimum",
			javaVersion:  "8.0.1",
			minVersion:   "11",
			maxVersion:   "17",
			expectedBool: false,
		},
		{
			name:         "Java version above maximum",
			javaVersion:  "18.0.1",
			minVersion:   "11",
			maxVersion:   "17",
			expectedBool: false,
		},
		{
			name:         "Java version equal to minimum",
			javaVersion:  "11",
			minVersion:   "11",
			maxVersion:   "17",
			expectedBool: true,
		},
		{
			name:         "Java version equal to maximum",
			javaVersion:  "17",
			minVersion:   "11",
			maxVersion:   "17",
			expectedBool: true,
		},
		{
			name:         "No min/max restrictions",
			javaVersion:  "8",
			minVersion:   "",
			maxVersion:   "",
			expectedBool: true,
		},
		{
			name:         "Only min restriction",
			javaVersion:  "11",
			minVersion:   "11",
			maxVersion:   "",
			expectedBool: true,
		},
		{
			name:         "Only max restriction",
			javaVersion:  "11",
			minVersion:   "",
			maxVersion:   "17",
			expectedBool: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isJavaVersionValid(tt.javaVersion, tt.minVersion, tt.maxVersion)
			assert.Equal(t, tt.expectedBool, result)
		})
	}
}

func TestBuildJavaCommand(t *testing.T) {
	// Create temp directory for testing
	tempDir, err := os.MkdirTemp("", "fg-jdk-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a mock JAR path
	jarPath := filepath.Join(tempDir, "test.jar")
	require.NoError(t, os.WriteFile(jarPath, []byte("mock jar"), 0644))

	// Define a mock server config
	serverCfg := config.ServerConfig{
		Port:      8080,
		Host:      "localhost",
		Contexts:  []string{"/api", "/fhir"},
		MaxMemory: "1g",
		Env: map[string]string{
			"DB_URL":  "jdbc:h2:mem:test",
			"LOG_DIR": "/var/log",
		},
	}

	// Test with all configuration options
	cmd, err := buildJavaCommand(jarPath, serverCfg, []string{"-XX:+UseG1GC"})
	assert.NoError(t, err)
	assert.Contains(t, cmd.Args, "java")
	assert.Contains(t, cmd.Args, "-Xmx1g")
	assert.Contains(t, cmd.Args, "-XX:+UseG1GC")
	assert.Contains(t, cmd.Args, "-jar")
	assert.Contains(t, cmd.Args, jarPath)
	assert.Contains(t, cmd.Args, "--server.port=8080")
	
	// Check environment variables
	foundDB := false
	foundLog := false
	for _, env := range cmd.Env {
		if env == "DB_URL=jdbc:h2:mem:test" {
			foundDB = true
		}
		if env == "LOG_DIR=/var/log" {
			foundLog = true
		}
	}
	assert.True(t, foundDB, "DB_URL environment variable should be set")
	assert.True(t, foundLog, "LOG_DIR environment variable should be set")

	// Test minimal configuration
	minServerCfg := config.ServerConfig{
		Port: 9090,
	}
	cmd, err = buildJavaCommand(jarPath, minServerCfg, nil)
	assert.NoError(t, err)
	assert.Contains(t, cmd.Args, "--server.port=9090")
}

func TestFindPIDByPort(t *testing.T) {
	// Skip on Windows as the implementation is platform-specific
	if os.PathSeparator == '\\' {
		t.Skip("Skipping on Windows as this test is platform-specific")
	}
	
	// Skip test on systems where lsof might not be available
	if _, err := os.Stat("/usr/bin/lsof"); os.IsNotExist(err) {
		t.Skip("lsof command not available, skipping test")
	}

	// This is a basic test and might be flaky depending on the system state
	// Real implementation should use a more reliable method or mock this function
	port := 0 // Using port 0 which shouldn't be in use
	pid, err := findPIDByPort(port)
	assert.Error(t, err) // Expect error since port 0 shouldn't be in use
	assert.Equal(t, 0, pid)
}

func TestWriteAndReadPIDFile(t *testing.T) {
	// Create temp directory for testing
	tempDir, err := os.MkdirTemp("", "fg-pid-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Test writing PID
	pidFile := filepath.Join(tempDir, "test.pid")
	testPID := 12345
	err = writePIDFile(pidFile, testPID)
	assert.NoError(t, err)

	// Test reading PID
	readPID, err := readPIDFile(pidFile)
	assert.NoError(t, err)
	assert.Equal(t, testPID, readPID)

	// Test reading non-existent PID file
	nonExistentPID, err := readPIDFile(filepath.Join(tempDir, "nonexistent.pid"))
	assert.Error(t, err)
	assert.Equal(t, 0, nonExistentPID)
}

func TestCheckIfProcessRunning(t *testing.T) {
	// This test will check if the current process is running (which it should be)
	currentPID := os.Getpid()
	assert.True(t, isProcessRunning(currentPID))

	// Check a PID that's very unlikely to exist
	assert.False(t, isProcessRunning(999999999))
}

func TestStartAndStopProcess(t *testing.T) {
	// Skip in CI environments
	if os.Getenv("CI") != "" {
		t.Skip("Skipping in CI environment")
	}

	// Check if we're running on Windows
	isWindows := os.PathSeparator == '\\'
	
	// Create temp directory for testing
	tempDir, err := os.MkdirTemp("", "fg-process-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	var cmd *exec.Cmd
	
	if isWindows {
		// On Windows, use a command that exists and will run for a while
		cmd, err = startProcess("ping", []string{"127.0.0.1", "-n", "10"}, map[string]string{})
	} else {
		// On Unix systems, we can use the sleep command
		scriptPath := filepath.Join(tempDir, "sleep.sh")
		scriptContent := `#!/bin/sh
echo "Starting test process"
sleep 10
`
		err = os.WriteFile(scriptPath, []byte(scriptContent), 0755)
		require.NoError(t, err)
		
		cmd, err = startProcess(scriptPath, []string{}, map[string]string{})
	}
	
	assert.NoError(t, err)
	assert.NotNil(t, cmd)
	assert.NotEqual(t, 0, cmd.Process.Pid)

	// Check if the process is running
	assert.True(t, isProcessRunning(cmd.Process.Pid))

	// Write PID file
	pidFile := filepath.Join(tempDir, "test.pid")
	err = writePIDFile(pidFile, cmd.Process.Pid)
	assert.NoError(t, err)

	// Stop the process
	err = stopProcess(int32(cmd.Process.Pid), false, 30)
	assert.NoError(t, err)

	// Wait a moment for the process to stop
	time.Sleep(100 * time.Millisecond)

	// Check if the process is still running
	assert.False(t, isProcessRunning(cmd.Process.Pid))
} 