package process

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/douglasedurocha/fg/pkg/util"
)

// ProcessInfo represents information about a running process
type ProcessInfo struct {
	PID         int       `json:"pid"`
	Version     string    `json:"version"`
	StartTime   time.Time `json:"start_time"`
	LogFile     string    `json:"log_file"`
	JDKPath     string    `json:"jdk_path"`
	CommandLine string    `json:"command_line"`
}

// ProcessManager manages running processes
type ProcessManager struct {
	pidFile string
	logsDir string
}

// NewProcessManager creates a new process manager
func NewProcessManager(pidFile, logsDir string) *ProcessManager {
	return &ProcessManager{
		pidFile: pidFile,
		logsDir: logsDir,
	}
}

// SaveProcess adds a new process to the track list
func (pm *ProcessManager) SaveProcess(info ProcessInfo) error {
	processes, err := pm.ListProcesses()
	if err != nil {
		processes = []ProcessInfo{}
	}

	// Add the new process
	processes = append(processes, info)

	// Save to file
	data, err := json.MarshalIndent(processes, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal process data: %w", err)
	}

	if err := os.WriteFile(pm.pidFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write process file: %w", err)
	}

	return nil
}

// ListProcesses returns a list of all tracked processes
func (pm *ProcessManager) ListProcesses() ([]ProcessInfo, error) {
	// Check if the file exists
	if _, err := os.Stat(pm.pidFile); os.IsNotExist(err) {
		return []ProcessInfo{}, nil
	}

	// Read the file
	data, err := os.ReadFile(pm.pidFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read process file: %w", err)
	}

	// Unmarshal the data
	var processes []ProcessInfo
	if err := json.Unmarshal(data, &processes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal process data: %w", err)
	}

	// Filter out dead processes
	liveProcesses := []ProcessInfo{}
	for _, process := range processes {
		if isProcessRunning(process.PID) {
			liveProcesses = append(liveProcesses, process)
		}
	}

	// If we filtered out some processes, update the file
	if len(liveProcesses) != len(processes) {
		data, _ := json.MarshalIndent(liveProcesses, "", "  ")
		os.WriteFile(pm.pidFile, data, 0644)
	}

	return liveProcesses, nil
}

// GetProcess returns information about a specific process
func (pm *ProcessManager) GetProcess(pid int) (*ProcessInfo, error) {
	processes, err := pm.ListProcesses()
	if err != nil {
		return nil, err
	}

	for _, process := range processes {
		if process.PID == pid {
			return &process, nil
		}
	}

	return nil, fmt.Errorf("process with PID %d not found", pid)
}

// RemoveProcess removes a process from the track list
func (pm *ProcessManager) RemoveProcess(pid int) error {
	processes, err := pm.ListProcesses()
	if err != nil {
		return err
	}

	newProcesses := []ProcessInfo{}
	for _, process := range processes {
		if process.PID != pid {
			newProcesses = append(newProcesses, process)
		}
	}

	// Save to file
	data, err := json.MarshalIndent(newProcesses, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal process data: %w", err)
	}

	if err := os.WriteFile(pm.pidFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write process file: %w", err)
	}

	return nil
}

// StopProcess stops a running process
func (pm *ProcessManager) StopProcess(pid int) error {
	// Check if the process exists
	if _, err := pm.GetProcess(pid); err != nil {
		return err
	}

	// Stop the process
	if err := util.StopProcess(pid); err != nil {
		return err
	}

	// Remove from tracking
	return pm.RemoveProcess(pid)
}

// GetLogPath returns the log file path for a specific version
func (pm *ProcessManager) GetLogPath(version string, pid int) string {
	return filepath.Join(pm.logsDir, fmt.Sprintf("%s_%d.log", version, pid))
}

// Helper function to check if a process is running
func isProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// On Windows, FindProcess always succeeds, so we need to check Signal
	// Signal 0 is used to check if a process exists
	err = process.Signal(os.Signal(nil))
	return err == nil
} 