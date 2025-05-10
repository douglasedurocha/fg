package download

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// File downloads a file from the specified URL and saves it to the destination
func File(url, destination string) error {
	// Create the directory if it doesn't exist
	dir := filepath.Dir(destination)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Create the file
	out, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", destination, err)
	}
	defer out.Close()

	// Send HTTP GET request
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download from %s: %w", url, err)
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad response status: %s", resp.Status)
	}

	// Write the response to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %w", destination, err)
	}

	return nil
}

// ExtractZip extracts a zip file to the specified destination
func ExtractZip(zipFile, destination string) error {
	// Open the zip file
	reader, err := zip.OpenReader(zipFile)
	if err != nil {
		return fmt.Errorf("failed to open zip file %s: %w", zipFile, err)
	}
	defer reader.Close()

	// Create destination directory
	if err := os.MkdirAll(destination, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", destination, err)
	}

	// Extract files
	for _, file := range reader.File {
		err := extractFile(file, destination)
		if err != nil {
			return fmt.Errorf("failed to extract file %s: %w", file.Name, err)
		}
	}

	return nil
}

// ExtractTarGz extracts a tar.gz file to the specified destination
func ExtractTarGz(tarGzFile, destination string) error {
	// For simplicity, we'll use external tar command for tar.gz files
	// This is less portable but much simpler than implementing tar.gz extraction in Go
	cmd := ""
	if strings.HasSuffix(tarGzFile, ".tar.gz") || strings.HasSuffix(tarGzFile, ".tgz") {
		cmd = fmt.Sprintf("mkdir -p %s && tar -xzf %s -C %s", destination, tarGzFile, destination)
	} else {
		return fmt.Errorf("unsupported archive format: %s", tarGzFile)
	}

	// Run the command
	if err := runCommand(cmd); err != nil {
		return fmt.Errorf("failed to extract archive: %w", err)
	}

	return nil
}

// Helper function to run a command
func runCommand(cmd string) error {
	// Split the command and arguments
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	// Create the command
	command := exec.Command(parts[0], parts[1:]...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	// Run the command
	return command.Run()
}

// Helper function to extract a single file from a zip archive
func extractFile(file *zip.File, destination string) error {
	// Prepare file path
	filePath := filepath.Join(destination, file.Name)

	// Check for path traversal
	if !strings.HasPrefix(filePath, filepath.Clean(destination)+string(os.PathSeparator)) {
		return fmt.Errorf("illegal file path: %s", filePath)
	}

	// Create directory if needed
	if file.FileInfo().IsDir() {
		if err := os.MkdirAll(filePath, file.Mode()); err != nil {
			return err
		}
		return nil
	}

	// Make sure the directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}

	// Create the file
	destFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Open source file
	srcFile, err := file.Open()
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Copy contents
	_, err = io.Copy(destFile, srcFile)
	return err
} 