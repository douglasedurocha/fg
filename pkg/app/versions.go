package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// GithubRelease represents a GitHub release
type GithubRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	PublishedAt time.Time `json:"published_at"`
	Assets      []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// ReleaseInfo stores version and release date information
type ReleaseInfo struct {
	Version     string
	ReleaseDate time.Time
}

// GetInstalledVersions returns a list of installed versions
func GetInstalledVersions() ([]string, error) {
	config := GetConfig()
	
	// Check if versions directory exists
	if _, err := os.Stat(config.VersionsDir); os.IsNotExist(err) {
		return []string{}, nil
	}
	
	// List directories in the versions directory
	files, err := os.ReadDir(config.VersionsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read versions directory: %w", err)
	}
	
	// Filter for directories
	versions := []string{}
	for _, file := range files {
		if file.IsDir() {
			versions = append(versions, file.Name())
		}
	}
	
	// Sort versions
	sort.Strings(versions)
	
	return versions, nil
}

// GetAvailableVersions returns a list of versions available for download
func GetAvailableVersions() ([]string, error) {
	releases, err := GetReleaseDetails()
	if err != nil {
		return nil, err
	}
	
	versions := make([]string, len(releases))
	for i, release := range releases {
		versions[i] = release.Version
	}
	
	return versions, nil
}

// GetReleaseDetails returns detailed information about available releases
func GetReleaseDetails() ([]ReleaseInfo, error) {
	// Fetch releases from GitHub
	url := fmt.Sprintf(GithubAPIFormat, GithubRepo)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()
	
	// Check response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad response status: %s", resp.Status)
	}
	
	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	
	// Unmarshal the response
	var releases []GithubRelease
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, fmt.Errorf("failed to unmarshal releases: %w", err)
	}
	
	// Extract version numbers and release dates
	releaseInfos := make([]ReleaseInfo, len(releases))
	for i, release := range releases {
		version := strings.TrimPrefix(release.TagName, "v")
		releaseInfos[i] = ReleaseInfo{
			Version:     version,
			ReleaseDate: release.PublishedAt,
		}
	}
	
	// Sort versions in descending order (newest first)
	sort.Slice(releaseInfos, func(i, j int) bool {
		return releaseInfos[i].ReleaseDate.After(releaseInfos[j].ReleaseDate)
	})
	
	return releaseInfos, nil
}

// GetLatestVersion returns the latest available version
func GetLatestVersion() (string, error) {
	releases, err := GetReleaseDetails()
	if err != nil {
		return "", err
	}
	
	if len(releases) == 0 {
		return "", fmt.Errorf("no versions available")
	}
	
	// The first version in the sorted list is the latest
	return releases[0].Version, nil
}

// IsVersionInstalled checks if a version is installed
func IsVersionInstalled(version string) bool {
	config := GetConfig()
	versionDir := filepath.Join(config.VersionsDir, version)
	_, err := os.Stat(versionDir)
	return err == nil
}

// GetVersionManifest reads the manifest for a specific version
func GetVersionManifest(version string) (*Manifest, error) {
	config := GetConfig()
	
	// Check if version is installed
	if !IsVersionInstalled(version) {
		return nil, fmt.Errorf("version %s is not installed", version)
	}
	
	versionDir := filepath.Join(config.VersionsDir, version)
	
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
		return nil, fmt.Errorf("manifest file not found for version %s", version)
	}
	
	// Read the manifest file
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}
	
	// Unmarshal the manifest
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal manifest: %w", err)
	}
	
	return &manifest, nil
} 