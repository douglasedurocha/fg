package app

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/douglasedurocha/fg/pkg/process"
)

// GetProcessManager returns a new process manager
func GetProcessManager() *process.ProcessManager {
	config := GetConfig()
	return process.NewProcessManager(config.PIDFile, config.LogsDir)
}

// List shows all installed versions
func List() error {
	// Initialize directories
	if err := InitDirectories(); err != nil {
		return err
	}

	installedVersions, err := GetInstalledVersions()
	if err != nil {
		return fmt.Errorf("failed to get installed versions: %w", err)
	}

	if len(installedVersions) == 0 {
		fmt.Println("No versions installed")
		return nil
	}

	fmt.Println("Installed versions:")
	config := GetConfig()
	for _, version := range installedVersions {
		versionDir := filepath.Join(config.VersionsDir, version)
		fmt.Printf("  %s [%s]\n", version, versionDir)
	}

	return nil
}

// Config shows the configuration for a specific version
func Config(version string) error {
	if !IsVersionInstalled(version) {
		return fmt.Errorf("version %s is not installed", version)
	}

	manifest, err := GetVersionManifest(version)
	if err != nil {
		return err
	}

	// Pretty print manifest
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	fmt.Printf("Configuration for version %s:\n%s\n", version, string(data))
	return nil
}

// getJavaPath returns the path to the java executable for a given JDK version
func getJavaPath(jdkVersion string) (string, error) {
	config := GetConfig()
	jdkDir := filepath.Join(config.JDKDir, jdkVersion)

	// Check OS specific paths
	var javaExe string
	if runtime.GOOS == "windows" {
		javaExe = filepath.Join(jdkDir, "bin", "java.exe")
	} else {
		javaExe = filepath.Join(jdkDir, "bin", "java")
	}

	if _, err := os.Stat(javaExe); err != nil {
		return "", fmt.Errorf("java executable not found at %s: %w", javaExe, err)
	}

	return javaExe, nil
}

// Start starts the application for a specific version
func Start(version string) error {
	// If no version specified, use the latest installed
	if version == "" {
		installedVersions, err := GetInstalledVersions()
		if err != nil {
			return fmt.Errorf("failed to get installed versions: %w", err)
		}

		if len(installedVersions) == 0 {
			return fmt.Errorf("no versions installed, please install one first")
		}

		version = installedVersions[len(installedVersions)-1]
		fmt.Printf("No version specified, using latest installed: %s\n", version)
	}

	if !IsVersionInstalled(version) {
		return fmt.Errorf("version %s is not installed", version)
	}

	manifest, err := GetVersionManifest(version)
	if err != nil {
		return err
	}

	// Get Java path
	javaPath, err := getJavaPath(manifest.JDK.Version)
	if err != nil {
		return err
	}

	config := GetConfig()
	versionDir := filepath.Join(config.VersionsDir, version)
	// Check for the app directory that might contain the actual JAR
	appDir := filepath.Join(versionDir, fmt.Sprintf("java-app-%s", version))
	if _, err := os.Stat(appDir); err == nil {
		// If the app directory exists, use it for execution
		versionDir = appDir
	}

	// Parse run command
	cmdParts := strings.Split(manifest.RunCommand, " ")
	args := []string{}
	for i, part := range cmdParts {
		if i == 0 && part == "java" {
			continue // Skip 'java' as we'll use the full path
		} else {
			// Replace jar path with full path if needed
			if strings.HasSuffix(part, ".jar") && !filepath.IsAbs(part) {
				// First look in the version directory
				if _, err := os.Stat(filepath.Join(versionDir, part)); err == nil {
					part = filepath.Join(versionDir, part)
				} else {
					// If not found directly, try common locations
					possibleJarPaths := []string{
						filepath.Join(versionDir, part),
						filepath.Join(versionDir, fmt.Sprintf("java-app-%s", version), part),
					}
					
					for _, jarPath := range possibleJarPaths {
						if _, err := os.Stat(jarPath); err == nil {
							part = jarPath
							break
						}
					}
				}
			}
			args = append(args, part)
		}
	}

	// Set up log file
	processManager := GetProcessManager()
	pid := time.Now().UnixNano() // Use timestamp as temp PID
	logFile := processManager.GetLogPath(version, int(pid))
	
	// Create log directory if it doesn't exist
	logDir := filepath.Dir(logFile)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file
	logWriter, err := os.Create(logFile)
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}
	defer logWriter.Close()

	// Start the process
	fmt.Printf("Starting java-app version %s...\n", version)
	
	// Create command
	cmd := exec.Command(javaPath, args...)
	cmd.Dir = versionDir
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter
	
	// Set environment variables
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("JAVA_HOME=%s", filepath.Dir(filepath.Dir(javaPath))))
	
	// Start the process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start process: %w", err)
	}

	pid = int64(cmd.Process.Pid)
	
	// Update log file name with actual PID
	newLogFile := processManager.GetLogPath(version, int(pid))
	logWriter.Close()
	if err := os.Rename(logFile, newLogFile); err != nil {
		fmt.Printf("Warning: failed to rename log file: %v\n", err)
		// Continue anyway, not a fatal error
	}

	// Save process info
	processInfo := process.ProcessInfo{
		PID:         int(pid),
		Version:     version,
		StartTime:   time.Now(),
		LogFile:     newLogFile,
		JDKPath:     javaPath,
		CommandLine: strings.Join(append([]string{javaPath}, args...), " "),
	}

	if err := processManager.SaveProcess(processInfo); err != nil {
		fmt.Printf("Warning: failed to save process info: %v\n", err)
		// Continue anyway, not a fatal error
	}

	fmt.Printf("Application started with PID: %d\n", pid)
	return nil
}

// Stop stops a running instance of the application
func Stop(pid string) error {
	processManager := GetProcessManager()

	// If no PID specified, stop all running instances
	if pid == "" {
		processes, err := processManager.ListProcesses()
		if err != nil {
			return fmt.Errorf("failed to list processes: %w", err)
		}

		if len(processes) == 0 {
			fmt.Println("No running instances found")
			return nil
		}

		fmt.Println("Stopping all running instances...")
		for _, proc := range processes {
			if err := processManager.StopProcess(proc.PID); err != nil {
				fmt.Printf("Failed to stop process %d: %v\n", proc.PID, err)
			} else {
				fmt.Printf("Stopped process with PID: %d\n", proc.PID)
			}
		}

		return nil
	}

	// Stop specific instance
	pidInt, err := strconv.Atoi(pid)
	if err != nil {
		return fmt.Errorf("invalid PID: %s", pid)
	}

	if err := processManager.StopProcess(pidInt); err != nil {
		return err
	}

	fmt.Printf("Stopped process with PID: %d\n", pidInt)
	return nil
}

// Status shows the status of all running instances
func Status() error {
	processManager := GetProcessManager()
	processes, err := processManager.ListProcesses()
	if err != nil {
		return fmt.Errorf("failed to list processes: %w", err)
	}

	if len(processes) == 0 {
		fmt.Println("No running instances found")
		return nil
	}

	fmt.Println("Running instances:")
	for _, proc := range processes {
		uptime := time.Since(proc.StartTime).Round(time.Second)
		fmt.Printf("  PID: %d | Version: %s | Uptime: %s\n", proc.PID, proc.Version, uptime)
	}

	return nil
}

// Logs shows the logs for a running instance
func Logs(pid string) error {
	processManager := GetProcessManager()

	// If no PID specified, show logs for the most recent instance
	if pid == "" {
		processes, err := processManager.ListProcesses()
		if err != nil {
			return fmt.Errorf("failed to list processes: %w", err)
		}

		if len(processes) == 0 {
			return fmt.Errorf("no running instances found")
		}

		// Use the most recent instance (assuming the last one in the list is the most recent)
		pid = strconv.Itoa(processes[len(processes)-1].PID)
		fmt.Printf("No PID specified, showing logs for most recent instance (PID: %s)\n", pid)
	}

	pidInt, err := strconv.Atoi(pid)
	if err != nil {
		return fmt.Errorf("invalid PID: %s", pid)
	}

	// Get process info
	proc, err := processManager.GetProcess(pidInt)
	if err != nil {
		// Process might not be running, but we can still try to find the log file
		// Check in the logs directory for a file matching the pattern
		config := GetConfig()
		logFiles, err := filepath.Glob(filepath.Join(config.LogsDir, fmt.Sprintf("*_%d.log", pidInt)))
		if err != nil || len(logFiles) == 0 {
			return fmt.Errorf("no logs found for PID %d", pidInt)
		}
		
		// Use the first log file found
		return showLogFile(logFiles[0])
	}

	return showLogFile(proc.LogFile)
}

// Helper function to show log file contents
func showLogFile(logFile string) error {
	file, err := os.Open(logFile)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	fmt.Printf("Showing logs from: %s\n", logFile)
	fmt.Println(strings.Repeat("-", 80))

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading log file: %w", err)
	}

	return nil
}

// Available shows all available versions with their release dates
func Available() error {
	// Initialize directories
	if err := InitDirectories(); err != nil {
		return err
	}

	releases, err := GetReleaseDetails()
	if err != nil {
		return fmt.Errorf("failed to get available versions: %w", err)
	}

	if len(releases) == 0 {
		fmt.Println("No versions available")
		return nil
	}

	fmt.Println("Version     Release Date")
	fmt.Println("--------    ------------")
	
	for _, release := range releases {
		fmt.Printf("%-12s%s\n", release.Version, release.ReleaseDate.Format("2006-01-02"))
	}

	return nil
} 