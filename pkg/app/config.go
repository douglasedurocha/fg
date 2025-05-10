package app

import (
	"os"
	"path/filepath"
	"runtime"
)

// Constants for the application
const (
	AppName         = "fg"
	GithubRepo      = "douglasedurocha/java-app"
	GithubAPIFormat = "https://api.github.com/repos/%s/releases"
	ReleasesURL     = "https://github.com/douglasedurocha/java-app/releases/download/v%s/java-app-%s.zip"
)

// AppConfig holds application configuration
type AppConfig struct {
	HomeDir      string
	VersionsDir  string
	JDKDir       string
	LogsDir      string
	PIDFile      string
	DownloadDir  string
	CurrentOSKey string
}

// GetConfig returns the application configuration
func GetConfig() *AppConfig {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	fgHome := filepath.Join(homeDir, "."+AppName)
	versionsDir := filepath.Join(fgHome, "versions")
	jdkDir := filepath.Join(fgHome, "jdk")
	logsDir := filepath.Join(fgHome, "logs")
	pidFile := filepath.Join(fgHome, "processes.json")
	downloadDir := filepath.Join(fgHome, "downloads")

	// Determine OS key for downloads
	osKey := "linux"
	switch runtime.GOOS {
	case "windows":
		osKey = "windows"
	case "darwin":
		osKey = "mac"
	}

	return &AppConfig{
		HomeDir:      fgHome,
		VersionsDir:  versionsDir,
		JDKDir:       jdkDir,
		LogsDir:      logsDir,
		PIDFile:      pidFile,
		DownloadDir:  downloadDir,
		CurrentOSKey: osKey,
	}
}

// InitDirectories ensures all required directories exist
func InitDirectories() error {
	cfg := GetConfig()
	dirs := []string{
		cfg.HomeDir,
		cfg.VersionsDir,
		cfg.JDKDir,
		cfg.LogsDir,
		cfg.DownloadDir,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
} 