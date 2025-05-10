package util

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// RunCommand executes a system command
func RunCommand(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ExecuteCommand runs a command and returns its output as string
func ExecuteCommand(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// ExtractArchive extracts an archive file to a destination directory
func ExtractArchive(archivePath, destDir string) error {
	// Ensure destination directory exists
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Handle based on file extension
	ext := strings.ToLower(filepath.Ext(archivePath))
	
	// For .zip files
	if ext == ".zip" {
		if runtime.GOOS == "windows" {
			// On Windows use PowerShell to extract
			return RunCommand("powershell", "-Command", fmt.Sprintf("Expand-Archive -Path '%s' -DestinationPath '%s' -Force", archivePath, destDir))
		} else {
			// On Unix-like systems use unzip
			return RunCommand("unzip", "-o", archivePath, "-d", destDir)
		}
	}
	
	// For .tar.gz files
	if ext == ".gz" && strings.HasSuffix(strings.TrimSuffix(archivePath, ".gz"), ".tar") {
		return RunCommand("tar", "-xzf", archivePath, "-C", destDir)
	}

	return fmt.Errorf("unsupported archive format: %s", archivePath)
}

// StartProcess starts a process and returns its PID
func StartProcess(command string, args ...string) (int, error) {
	cmd := exec.Command(command, args...)
	
	// Set up stdout/stderr redirection to a log file if needed
	// This would be handled in a real implementation
	
	// Start the process
	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("failed to start process: %w", err)
	}
	
	return cmd.Process.Pid, nil
}

// StopProcess stops a process by its PID
func StopProcess(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("process with PID %d not found: %w", pid, err)
	}
	
	if err := process.Kill(); err != nil {
		return fmt.Errorf("failed to kill process %d: %w", pid, err)
	}
	
	return nil
} 