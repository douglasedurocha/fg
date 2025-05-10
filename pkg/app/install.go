package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/douglasedurocha/fg/pkg/download"
	"github.com/douglasedurocha/fg/pkg/util"
)

// Install downloads and installs a specific version of the application
func Install(version string) error {
	// Initialize directories
	if err := InitDirectories(); err != nil {
		return fmt.Errorf("failed to initialize directories: %w", err)
	}

	// Check if version is already installed
	if IsVersionInstalled(version) {
		return fmt.Errorf("version %s is already installed", version)
	}

	config := GetConfig()

	// Create version directory
	versionDir := filepath.Join(config.VersionsDir, version)
	if err := os.MkdirAll(versionDir, 0755); err != nil {
		return fmt.Errorf("failed to create version directory: %w", err)
	}

	// Download application
	fmt.Printf("Downloading java-app-%s.zip...\n", version)
	zipURL := fmt.Sprintf(ReleasesURL, version, version)
	zipFile := filepath.Join(config.DownloadDir, fmt.Sprintf("java-app-%s.zip", version))
	
	if err := download.File(zipURL, zipFile); err != nil {
		return fmt.Errorf("failed to download application: %w", err)
	}

	// Extract the zip file
	fmt.Printf("Extracting java-app-%s.zip...\n", version)
	if err := util.ExtractArchive(zipFile, versionDir); err != nil {
		return fmt.Errorf("failed to extract application: %w", err)
	}

	// Look for the manifest file in different possible locations
	manifestPaths := []string{
		filepath.Join(versionDir, "fgmanifest"),
		filepath.Join(versionDir, "fgmanifest.json"),
		filepath.Join(versionDir, fmt.Sprintf("java-app-%s", version), "fgmanifest"),
		filepath.Join(versionDir, fmt.Sprintf("java-app-%s", version), "fgmanifest.json"),
	}

	var manifestPath string
	var manifestFound bool

	for _, path := range manifestPaths {
		if _, err := os.Stat(path); err == nil {
			manifestPath = path
			manifestFound = true
			break
		}
	}

	if !manifestFound {
		return fmt.Errorf("manifest file not found in the extracted files. Looked in: %v", manifestPaths)
	}

	// Read the manifest
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read manifest file: %w", err)
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("failed to unmarshal manifest: %w", err)
	}

	// If manifest is in a subdirectory, copy it to the version root
	if filepath.Dir(manifestPath) != versionDir {
		rootManifestPath := filepath.Join(versionDir, "fgmanifest")
		if err := os.WriteFile(rootManifestPath, data, 0644); err != nil {
			return fmt.Errorf("failed to copy manifest to root directory: %w", err)
		}
	}

	// Install JDK if needed
	if err := InstallJDK(&manifest); err != nil {
		return fmt.Errorf("failed to install JDK: %w", err)
	}

	// Download dependencies if needed
	if err := DownloadDependencies(&manifest, versionDir); err != nil {
		return fmt.Errorf("failed to download dependencies: %w", err)
	}

	fmt.Printf("Successfully installed version %s\n", version)
	return nil
}

// InstallJDK installs the JDK required by the manifest
func InstallJDK(manifest *Manifest) error {
	config := GetConfig()
	jdkVersion := manifest.JDK.Version
	jdkDir := filepath.Join(config.JDKDir, jdkVersion)

	// Check if JDK is already installed
	if _, err := os.Stat(jdkDir); err == nil {
		fmt.Printf("JDK %s is already installed\n", jdkVersion)
		return nil
	}

	// Get download URL based on OS
	downloadURL, ok := manifest.JDK.Download[config.CurrentOSKey]
	if !ok {
		return fmt.Errorf("no JDK download URL for %s", config.CurrentOSKey)
	}

	fmt.Printf("Downloading JDK %s...\n", jdkVersion)

	// Determine file extension
	var ext string
	if strings.HasSuffix(downloadURL, ".zip") {
		ext = ".zip"
	} else if strings.HasSuffix(downloadURL, ".tar.gz") {
		ext = ".tar.gz"
	} else {
		return fmt.Errorf("unsupported JDK archive format: %s", downloadURL)
	}

	// Download JDK
	jdkFile := filepath.Join(config.DownloadDir, fmt.Sprintf("jdk%s%s", jdkVersion, ext))
	if err := download.File(downloadURL, jdkFile); err != nil {
		return fmt.Errorf("failed to download JDK: %w", err)
	}

	// Create JDK directory
	if err := os.MkdirAll(jdkDir, 0755); err != nil {
		return fmt.Errorf("failed to create JDK directory: %w", err)
	}

	// Extract JDK
	fmt.Printf("Extracting JDK %s...\n", jdkVersion)
	if err := util.ExtractArchive(jdkFile, jdkDir); err != nil {
		return fmt.Errorf("failed to extract JDK: %w", err)
	}

	// JDK archives often have a top-level directory, we need to find it
	files, err := os.ReadDir(jdkDir)
	if err != nil {
		return fmt.Errorf("failed to read JDK directory: %w", err)
	}

	// If there's only one directory and it's a JDK directory, move its contents up
	if len(files) == 1 && files[0].IsDir() && strings.Contains(files[0].Name(), "jdk") {
		subdir := filepath.Join(jdkDir, files[0].Name())
		tempDir := filepath.Join(config.JDKDir, fmt.Sprintf("temp-%s", jdkVersion))
		
		// Move to temp directory first
		if err := os.Rename(subdir, tempDir); err != nil {
			return fmt.Errorf("failed to move JDK directory: %w", err)
		}
		
		// Remove original directory
		if err := os.RemoveAll(jdkDir); err != nil {
			return fmt.Errorf("failed to remove JDK directory: %w", err)
		}
		
		// Move from temp to original
		if err := os.Rename(tempDir, jdkDir); err != nil {
			return fmt.Errorf("failed to move JDK from temp: %w", err)
		}
	}

	fmt.Printf("Successfully installed JDK %s\n", jdkVersion)
	return nil
}

// DownloadDependencies downloads the dependencies specified in the manifest
func DownloadDependencies(manifest *Manifest, versionDir string) error {
	if len(manifest.Dependencies) == 0 {
		return nil
	}

	fmt.Println("Downloading dependencies...")
	libDir := filepath.Join(versionDir, "lib")
	if err := os.MkdirAll(libDir, 0755); err != nil {
		return fmt.Errorf("failed to create lib directory: %w", err)
	}

	for _, dependency := range manifest.Dependencies {
		// For simplicity, we'll use Maven Central
		groupPath := strings.ReplaceAll(dependency.GroupID, ".", "/")
		fileName := fmt.Sprintf("%s-%s.jar", dependency.ArtifactID, dependency.Version)
		url := fmt.Sprintf("https://repo1.maven.org/maven2/%s/%s/%s/%s",
			groupPath, dependency.ArtifactID, dependency.Version, fileName)
		
		destPath := filepath.Join(libDir, fileName)
		fmt.Printf("Downloading %s...\n", fileName)
		
		if err := download.File(url, destPath); err != nil {
			return fmt.Errorf("failed to download dependency %s: %w", fileName, err)
		}
	}

	fmt.Println("Successfully downloaded all dependencies")
	return nil
}

// Uninstall removes an installed version
func Uninstall(version string) error {
	if !IsVersionInstalled(version) {
		return fmt.Errorf("version %s is not installed", version)
	}

	// Check if there are running instances of this version
	processManager := GetProcessManager()
	processes, err := processManager.ListProcesses()
	if err != nil {
		return fmt.Errorf("failed to list processes: %w", err)
	}

	for _, process := range processes {
		if process.Version == version {
			return fmt.Errorf("version %s has running instances, please stop them first", version)
		}
	}

	config := GetConfig()
	versionDir := filepath.Join(config.VersionsDir, version)
	
	fmt.Printf("Uninstalling version %s...\n", version)
	if err := os.RemoveAll(versionDir); err != nil {
		return fmt.Errorf("failed to remove version directory: %w", err)
	}

	fmt.Printf("Successfully uninstalled version %s\n", version)
	return nil
}

// Update installs the latest available version
func Update() error {
	latestVersion, err := GetLatestVersion()
	if err != nil {
		return fmt.Errorf("failed to get latest version: %w", err)
	}

	if IsVersionInstalled(latestVersion) {
		fmt.Printf("Latest version %s is already installed\n", latestVersion)
		return nil
	}

	fmt.Printf("Installing latest version (%s)...\n", latestVersion)
	return Install(latestVersion)
} 